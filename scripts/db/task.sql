drop table if exists `robber`.`task`;
create table `robber`.`task` (
    `date` VARCHAR(32) NOT NULL COMMENT '日期',
    `completed` TINYINT NOT NULL COMMENT '是否完成',
    `metadata_count` INT NOT NULL COMMENT '元数据量',
    `stock_count` INT NOT NULL COMMENT 'stock 数据量',
    `day_count` INT NOT NULL COMMENT 'day 数据量',
    `week_count` INT NOT NULL COMMENT 'week 数据量',
    `callback_url` TEXT NOT NULL COMMENT 'callback url',
    `create_timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `modify_timestamp` TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`date`)
);
