-- 创建配置历史表
CREATE TABLE IF NOT EXISTS `config_histories` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `config_id`   BIGINT UNSIGNED NOT NULL COMMENT '关联 configs.id',
  `namespace`   VARCHAR(64)     NOT NULL,
  `group`       VARCHAR(128)    NOT NULL,
  `data_id`     VARCHAR(256)    NOT NULL,
  `content`     MEDIUMTEXT      NOT NULL COMMENT '历史配置内容',
  `content_md5` CHAR(32)        NOT NULL,
  `version`     BIGINT UNSIGNED NOT NULL COMMENT '该历史对应的版本号',
  `op_type`     VARCHAR(16)     NOT NULL COMMENT '操作类型：create / update / delete',
  `op_by`       VARCHAR(64)     NOT NULL DEFAULT '',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_config_id` (`config_id`),
  KEY `idx_namespace_group_dataid` (`namespace`, `group`, `data_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置变更历史';