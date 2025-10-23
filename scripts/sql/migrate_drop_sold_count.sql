-- 注意：执行前请先备份表或在测试环境验证
-- 该迁移在代码已改为仅使用 remaining_count 后执行

ALTER TABLE flash_sale_activities DROP COLUMN sold_count;

