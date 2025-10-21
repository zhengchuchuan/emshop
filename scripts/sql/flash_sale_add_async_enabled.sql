ALTER TABLE flash_sale_activities
  ADD COLUMN IF NOT EXISTS async_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用异步链路(1异步,0同步)';

