-- create table quote_day
drop table if exists `robber`.`quote_day`;
create table `robber`.`quote_day` (
    `id` BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `code` CHAR(8) NOT NULL COMMENT '股票代码',
    `open` DECIMAL(10,2) NOT NULL COMMENT '开盘价',
    `close` DECIMAL(10,2) NOT NULL COMMENT '收盘价',
    `high` DECIMAL(10,2) NOT NULL COMMENT '最高价',
    `low` DECIMAL(10,2) NOT NULL COMMENT '最低价',
    `yesterday_closed` DECIMAL(10,2) NOT NULL COMMENT '昨日收盘价',
    `volume` BIGINT NOT NULL COMMENT '交易量',
    `account` DECIMAL(18,2) NOT NULL COMMENT '金额',
    `date` TIMESTAMP NOT NULL COMMENT '日期',
    `num_of_year` INT NOT NULL COMMENT '天数',
    `xd` DOUBLE NOT NULL COMMENT '前复权比例',
    `create_timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `modify_timestamp` TIMESTAMP COMMENT '修改时间'
);
create index idx_code_date on `robber`.`quote_day`(`code`,`date`);

drop table if exists `robber`.`quote_week`;
create table `robber`.`quote_week` (
    `id` BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `code` CHAR(8) NOT NULL COMMENT '股票代码',
    `open` DECIMAL(10,2) NOT NULL COMMENT '开盘价',
    `close` DECIMAL(10,2) NOT NULL COMMENT '收盘价',
    `high` DECIMAL(10,2) NOT NULL COMMENT '最高价',
    `low` DECIMAL(10,2) NOT NULL COMMENT '最低价',
    `yesterday_closed` DECIMAL(10,2) NOT NULL COMMENT '昨日收盘价',
    `volume` BIGINT NOT NULL COMMENT '交易量',
    `account` DECIMAL(18,2) NOT NULL COMMENT '金额',
    `date` TIMESTAMP NOT NULL COMMENT '开始时期',
    `num_of_year` INT NOT NULL COMMENT '周数',
    `xd` DOUBLE NOT NULL COMMENT '前复权比例',
    `create_timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `modify_timestamp` TIMESTAMP COMMENT '修改时间'
);
create index idx_code_date_end on `robber`.`quote_week`(`code`,`date`);
