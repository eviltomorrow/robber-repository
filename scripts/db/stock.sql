-- create table taskrecord
drop table if exists `robber`.`stock`;
create table `robber`.`stock` (
    `code` CHAR(8) NOT NULL COMMENT '股票代码',
    `name` VARCHAR(32) NOT NULL COMMENT '名称',
    `suspend` VARCHAR(32) NOT NULL COMMENT '停牌状态',
    `create_timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `modify_timestamp` TIMESTAMP COMMENT '修改时间',
     PRIMARY KEY(`code`)
);
