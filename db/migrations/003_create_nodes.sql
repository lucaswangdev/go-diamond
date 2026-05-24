-- 创建服务节点表
CREATE TABLE IF NOT EXISTS `server_nodes` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `node_id`    VARCHAR(64)     NOT NULL COMMENT '节点唯一ID（UUID）',
  `address`    VARCHAR(128)    NOT NULL COMMENT '节点地址 ip:port',
  `status`     VARCHAR(16)     NOT NULL DEFAULT 'online' COMMENT 'online / offline',
  `last_beat`  DATETIME        NOT NULL COMMENT '最后心跳时间',
  `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_node_id` (`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务节点注册表';