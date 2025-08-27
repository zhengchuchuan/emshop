package calculator

import (
	"fmt"
	"math"
	"sort"

	"emshop/pkg/log"
)

// CombinationOptimizer 优惠券组合优化器
type CombinationOptimizer struct {
	maxCombinations int
	strategy        OptimizationStrategy
}

// OptimizationStrategy 优化策略
type OptimizationStrategy string

const (
	StrategyMaxDiscount   OptimizationStrategy = "max_discount"   // 最大折扣优先
	StrategyMaxCoupons    OptimizationStrategy = "max_coupons"    // 最多优惠券优先
	StrategyBalanced      OptimizationStrategy = "balanced"       // 平衡策略
	StrategyGreedy        OptimizationStrategy = "greedy"         // 贪心策略
)

// CombinationResult 组合结果
type CombinationResult struct {
	Combination     []*CalculationResult
	TotalDiscount   float64
	CouponCount     int
	Score           float64
	Description     string
}

// NewCombinationOptimizer 创建优化器
func NewCombinationOptimizer() *CombinationOptimizer {
	return &CombinationOptimizer{
		maxCombinations: 32, // 最多检查32种组合
		strategy:        StrategyMaxDiscount,
	}
}

// SetStrategy 设置优化策略
func (co *CombinationOptimizer) SetStrategy(strategy OptimizationStrategy) {
	co.strategy = strategy
}

// OptimizeCombination 优化优惠券组合
func (co *CombinationOptimizer) OptimizeCombination(
	ctx *CalculationContext, 
	calculationResults []*CalculationResult,
) ([]*CalculationResult, error) {
	log.Infof("开始优化优惠券组合，候选结果数量: %d, 策略: %s", 
		len(calculationResults), co.strategy)

	// 过滤出可用的优惠券
	availableResults := make([]*CalculationResult, 0)
	for _, result := range calculationResults {
		if result.IsApplicable {
			availableResults = append(availableResults, result)
		}
	}

	if len(availableResults) == 0 {
		return calculationResults, nil
	}

	// 根据策略选择最优组合
	var bestCombination []*CalculationResult
	var err error

	switch co.strategy {
	case StrategyMaxDiscount:
		bestCombination, err = co.findMaxDiscountCombination(ctx, availableResults)
	case StrategyMaxCoupons:
		bestCombination, err = co.findMaxCouponsCombination(ctx, availableResults)
	case StrategyBalanced:
		bestCombination, err = co.findBalancedCombination(ctx, availableResults)
	case StrategyGreedy:
		bestCombination, err = co.findGreedyCombination(ctx, availableResults)
	default:
		bestCombination, err = co.findMaxDiscountCombination(ctx, availableResults)
	}

	if err != nil {
		return calculationResults, err
	}

	// 将最优组合结果合并回原结果中
	return co.mergeResults(calculationResults, bestCombination), nil
}

// findMaxDiscountCombination 找到最大折扣组合
func (co *CombinationOptimizer) findMaxDiscountCombination(
	ctx *CalculationContext,
	results []*CalculationResult,
) ([]*CalculationResult, error) {
	log.Info("执行最大折扣优化策略")

	// 生成所有可能的组合
	combinations := co.generateValidCombinations(ctx, results)
	if len(combinations) == 0 {
		return results, nil
	}

	// 找到总折扣最大的组合
	bestCombination := combinations[0]
	maxDiscount := bestCombination.TotalDiscount

	for _, combination := range combinations[1:] {
		if combination.TotalDiscount > maxDiscount {
			maxDiscount = combination.TotalDiscount
			bestCombination = combination
		}
	}

	log.Infof("最大折扣组合选择完成，折扣金额: %.2f, 优惠券数量: %d", 
		bestCombination.TotalDiscount, bestCombination.CouponCount)

	return bestCombination.Combination, nil
}

// findMaxCouponsCombination 找到最多优惠券组合
func (co *CombinationOptimizer) findMaxCouponsCombination(
	ctx *CalculationContext,
	results []*CalculationResult,
) ([]*CalculationResult, error) {
	log.Info("执行最多优惠券优化策略")

	combinations := co.generateValidCombinations(ctx, results)
	if len(combinations) == 0 {
		return results, nil
	}

	// 找到优惠券数量最多的组合，折扣相同时选择优惠券更多的
	bestCombination := combinations[0]
	
	for _, combination := range combinations[1:] {
		if combination.CouponCount > bestCombination.CouponCount ||
			(combination.CouponCount == bestCombination.CouponCount && 
			 combination.TotalDiscount > bestCombination.TotalDiscount) {
			bestCombination = combination
		}
	}

	log.Infof("最多优惠券组合选择完成，优惠券数量: %d, 折扣金额: %.2f", 
		bestCombination.CouponCount, bestCombination.TotalDiscount)

	return bestCombination.Combination, nil
}

// findBalancedCombination 找到平衡组合
func (co *CombinationOptimizer) findBalancedCombination(
	ctx *CalculationContext,
	results []*CalculationResult,
) ([]*CalculationResult, error) {
	log.Info("执行平衡优化策略")

	combinations := co.generateValidCombinations(ctx, results)
	if len(combinations) == 0 {
		return results, nil
	}

	// 计算平衡得分：折扣金额 + 优惠券数量 * 权重
	bestCombination := combinations[0]
	bestScore := co.calculateBalancedScore(bestCombination, ctx.OrderAmount)

	for _, combination := range combinations[1:] {
		score := co.calculateBalancedScore(combination, ctx.OrderAmount)
		if score > bestScore {
			bestScore = score
			bestCombination = combination
		}
	}

	log.Infof("平衡组合选择完成，得分: %.2f, 折扣金额: %.2f, 优惠券数量: %d", 
		bestScore, bestCombination.TotalDiscount, bestCombination.CouponCount)

	return bestCombination.Combination, nil
}

// findGreedyCombination 找到贪心组合
func (co *CombinationOptimizer) findGreedyCombination(
	ctx *CalculationContext,
	results []*CalculationResult,
) ([]*CalculationResult, error) {
	log.Info("执行贪心优化策略")

	// 按单张优惠券的折扣金额排序
	sortedResults := make([]*CalculationResult, len(results))
	copy(sortedResults, results)
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].DiscountAmount > sortedResults[j].DiscountAmount
	})

	// 贪心选择：从折扣最大的开始选择
	selectedResults := make([]*CalculationResult, 0)
	totalDiscount := 0.0
	remainingAmount := ctx.OrderAmount

	for _, result := range sortedResults {
		// 检查是否可以与已选择的优惠券组合
		tempCombination := append(selectedResults, result)
		if co.isValidCombination(ctx, tempCombination) {
			// 重新计算折扣，考虑组合效果
			adjustedDiscount := co.calculateAdjustedDiscount(result, remainingAmount)
			if adjustedDiscount > 0 {
				selectedResults = append(selectedResults, result)
				totalDiscount += adjustedDiscount
				remainingAmount -= adjustedDiscount
				
				if remainingAmount <= 0 {
					break
				}
			}
		}
	}

	log.Infof("贪心组合选择完成，折扣金额: %.2f, 优惠券数量: %d", 
		totalDiscount, len(selectedResults))

	return selectedResults, nil
}

// generateValidCombinations 生成有效组合
func (co *CombinationOptimizer) generateValidCombinations(
	ctx *CalculationContext,
	results []*CalculationResult,
) []*CombinationResult {
	combinations := make([]*CombinationResult, 0)
	n := len(results)
	
	// 限制组合数量避免组合爆炸
	maxResults := int(math.Min(float64(n), 6)) // 最多6张优惠券的组合
	
	// 生成所有可能的组合（使用位运算）
	maxCombination := 1 << uint(maxResults)
	if maxCombination > co.maxCombinations {
		maxCombination = co.maxCombinations
	}

	for i := 1; i < maxCombination; i++ {
		combination := make([]*CalculationResult, 0)
		
		// 根据位来选择优惠券
		for j := 0; j < maxResults && j < n; j++ {
			if (i>>uint(j))&1 == 1 {
				combination = append(combination, results[j])
			}
		}
		
		// 验证组合是否有效
		if co.isValidCombination(ctx, combination) {
			combinationResult := co.evaluateCombination(ctx, combination)
			combinations = append(combinations, combinationResult)
		}
	}

	log.Infof("生成了%d个有效组合", len(combinations))
	return combinations
}

// isValidCombination 检查组合是否有效
func (co *CombinationOptimizer) isValidCombination(
	ctx *CalculationContext, 
	combination []*CalculationResult,
) bool {
	// 检查优惠券数量限制
	if len(combination) > 5 { // 最多5张优惠券
		return false
	}

	// 简化的组合规则检查
	// 这里简化处理，实际中应该根据优惠券类型进行更精确的验证
	// 目前暂时允许所有组合，后续可以根据业务需求完善
	_ = combination // 避免未使用变量警告

	return true
}

// evaluateCombination 评估组合
func (co *CombinationOptimizer) evaluateCombination(
	ctx *CalculationContext,
	combination []*CalculationResult,
) *CombinationResult {
	totalDiscount := 0.0
	for _, result := range combination {
		totalDiscount += result.DiscountAmount
	}

	// 确保总折扣不超过订单金额
	totalDiscount = math.Min(totalDiscount, ctx.OrderAmount)

	return &CombinationResult{
		Combination:   combination,
		TotalDiscount: totalDiscount,
		CouponCount:   len(combination),
		Score:         co.calculateScore(totalDiscount, len(combination), ctx.OrderAmount),
		Description:   fmt.Sprintf("%d张优惠券，总折扣%.2f", len(combination), totalDiscount),
	}
}

// calculateBalancedScore 计算平衡得分
func (co *CombinationOptimizer) calculateBalancedScore(
	combination *CombinationResult,
	orderAmount float64,
) float64 {
	// 平衡得分 = 折扣金额权重 * 折扣金额 + 数量权重 * 优惠券数量
	discountWeight := 0.7
	countWeight := 0.3
	
	// 归一化处理
	normalizedDiscount := combination.TotalDiscount / orderAmount
	normalizedCount := float64(combination.CouponCount) / 5.0 // 假设最多5张优惠券
	
	return discountWeight*normalizedDiscount + countWeight*normalizedCount
}

// calculateAdjustedDiscount 计算调整后的折扣
func (co *CombinationOptimizer) calculateAdjustedDiscount(
	result *CalculationResult, 
	remainingAmount float64,
) float64 {
	// 根据剩余金额调整折扣
	return math.Min(result.DiscountAmount, remainingAmount)
}

// calculateScore 计算组合得分
func (co *CombinationOptimizer) calculateScore(
	totalDiscount float64, 
	couponCount int, 
	orderAmount float64,
) float64 {
	// 基础得分 = 折扣金额
	score := totalDiscount
	
	// 优惠券数量bonus（鼓励使用多张优惠券）
	score += float64(couponCount) * 2
	
	// 折扣率bonus
	discountRate := totalDiscount / orderAmount
	if discountRate > 0.3 { // 超过30%折扣的额外bonus
		score += totalDiscount * 0.1
	}
	
	return score
}

// mergeResults 合并结果
func (co *CombinationOptimizer) mergeResults(
	originalResults []*CalculationResult,
	optimizedResults []*CalculationResult,
) []*CalculationResult {
	// 创建优化后的优惠券ID集合
	optimizedIDs := make(map[int64]bool)
	for _, result := range optimizedResults {
		optimizedIDs[result.CouponID] = true
	}

	// 合并结果
	mergedResults := make([]*CalculationResult, 0, len(originalResults))
	
	for _, result := range originalResults {
		if optimizedIDs[result.CouponID] {
			// 使用优化后的结果
			for _, optimizedResult := range optimizedResults {
				if optimizedResult.CouponID == result.CouponID {
					mergedResults = append(mergedResults, optimizedResult)
					break
				}
			}
		} else {
			// 标记为未选中
			result.IsApplicable = false
			result.Reason = "未被选入最优组合"
			mergedResults = append(mergedResults, result)
		}
	}

	return mergedResults
}

// OptimizationReport 优化报告
type OptimizationReport struct {
	OriginalCouponCount    int
	OptimizedCouponCount   int
	OriginalTotalDiscount  float64
	OptimizedTotalDiscount float64
	ImprovementPercentage  float64
	Strategy               OptimizationStrategy
	ExecutionTime          int64 // 毫秒
}

// GenerateReport 生成优化报告
func (co *CombinationOptimizer) GenerateReport(
	originalResults []*CalculationResult,
	optimizedResults []*CalculationResult,
	executionTime int64,
) *OptimizationReport {
	originalDiscount := 0.0
	originalCount := 0
	for _, result := range originalResults {
		if result.IsApplicable {
			originalDiscount += result.DiscountAmount
			originalCount++
		}
	}

	optimizedDiscount := 0.0
	optimizedCount := 0
	for _, result := range optimizedResults {
		if result.IsApplicable {
			optimizedDiscount += result.DiscountAmount
			optimizedCount++
		}
	}

	improvement := 0.0
	if originalDiscount > 0 {
		improvement = (optimizedDiscount - originalDiscount) / originalDiscount * 100
	}

	return &OptimizationReport{
		OriginalCouponCount:    originalCount,
		OptimizedCouponCount:   optimizedCount,
		OriginalTotalDiscount:  originalDiscount,
		OptimizedTotalDiscount: optimizedDiscount,
		ImprovementPercentage:  improvement,
		Strategy:               co.strategy,
		ExecutionTime:          executionTime,
	}
}