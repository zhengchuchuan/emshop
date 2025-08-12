package flag

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/pflag"
)

// NamedFlagSets 命名标志集合管理器
// 这是项目CLI系统的核心组件，实现了标志的分组管理功能
// 解决了大型应用程序标志过多、难以组织的问题
type NamedFlagSets struct {
	// Order 存储标志集合名称的有序列表，用于控制帮助信息的显示顺序
	Order []string
	// FlagSets 存储各个命名的pflag.FlagSet，key为组名，value为对应的标志集合
	FlagSets map[string]*pflag.FlagSet
}

// FlagSet 获取或创建指定名称的标志集合
// 这是懒加载模式的实现：只有在需要时才创建pflag.FlagSet
// 同时维护显示顺序，确保帮助信息的一致性
func (nfs *NamedFlagSets) FlagSet(name string) *pflag.FlagSet {
	// 懒加载：第一次调用时初始化map
	if nfs.FlagSets == nil {
		nfs.FlagSets = map[string]*pflag.FlagSet{}
	}
	// 如果指定名称的标志集合不存在，则创建它
	if _, ok := nfs.FlagSets[name]; !ok {
		// 创建新的pflag.FlagSet，设置错误处理策略为ExitOnError
		nfs.FlagSets[name] = pflag.NewFlagSet(name, pflag.ExitOnError)
		// 记录创建顺序，用于后续的有序显示
		nfs.Order = append(nfs.Order, name)
	}
	return nfs.FlagSets[name]
}

// PrintSections 以分组形式打印所有标志集合的帮助信息
// 这是CLI帮助系统的核心实现，将多个标志组格式化输出
// 与Cobra的帮助系统集成，提供专业的用户体验
func PrintSections(w io.Writer, fss NamedFlagSets, cols int) {
	// 按照创建顺序遍历所有标志组
	for _, name := range fss.Order {
		fs := fss.FlagSets[name]
		// 跳过没有标志的空组
		if !fs.HasFlags() {
			continue
		}

		// 创建临时的FlagSet用于格式化输出
		// 这是为了利用pflag的内建格式化功能
		wideFS := pflag.NewFlagSet("", pflag.ExitOnError)
		wideFS.AddFlagSet(fs) // 复制原FlagSet的所有标志

		// 处理列宽限制的技巧：添加一个dummy标志来控制格式
		var zzz string
		if cols > 24 {
			zzz = strings.Repeat("z", cols-24)
			// 添加一个长名称的虚拟标志，用于控制输出宽度
			wideFS.Int(zzz, 0, strings.Repeat("z", cols-24))
		}

		// 生成格式化的标志帮助信息
		var buf bytes.Buffer
		// 标志组标题格式：首字母大写 + " flags:"
		fmt.Fprintf(&buf, "\n%s flags:\n\n%s", 
			strings.ToUpper(name[:1])+name[1:], // "server" -> "Server"
			wideFS.FlagUsagesWrapped(cols))     // pflag自动格式化标志列表

		// 移除之前添加的dummy标志，保持输出干净
		if cols > 24 {
			i := strings.Index(buf.String(), zzz)
			lines := strings.Split(buf.String()[:i], "\n")
			fmt.Fprint(w, strings.Join(lines[:len(lines)-1], "\n"))
			fmt.Fprintln(w)
		} else {
			fmt.Fprint(w, buf.String())
		}
	}
}
