# create database btc_database
CREATE DATABASE IF NOT EXISTS btc_database;

CREATE TABLE IF NOT EXISTS btc_database.t_block_info (
    `id`                            BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `height`                        INT                 NOT NULL DEFAULT 0      COMMENT '区块高度',
    `isfork`                        TINYINT             NOT NULL DEFAULT 0      COMMENT '是否为分叉链，0表示为主链，1表示为分叉链',
    `version`                       INT                 NOT NULL DEFAULT 0      COMMENT '版本号',
    `time`                          BIGINT              NOT NULL DEFAULT 0      COMMENT '区块打包时间',
    `bits`                          varchar(64)         NOT NULL DEFAULT ''     COMMENT '难度对应值',
    `nonce`                         BIGINT              NOT NULL DEFAULT 0      COMMENT '随机数',
    `difficulty`                    DOUBLE              NOT NULL DEFAULT 0      COMMENT '难度值',
    `size`                          INT                 NOT NULL DEFAULT 0      COMMENT '区块大小',
    `weight`                        BIGINT              NOT NULL DEFAULT 0      COMMENT '权重',
    `mediantime`                    BIGINT              NOT NULL DEFAULT 0      COMMENT '时间',
    `chainwork`                     varchar(64)         NOT NULL DEFAULT ''     COMMENT '哈希次数',
    `hash`                          CHAR(64)            NOT NULL DEFAULT ''     COMMENT '当前区块的哈希',
    `merkleroot`                    CHAR(64)            NOT NULL DEFAULT ''     COMMENT '默克尔树根',
    `previousblockhash`             CHAR(64)            NOT NULL DEFAULT ''     COMMENT '前一个区块哈希',
    `ntx`                           BIGINT              NOT NULL DEFAULT 0      COMMENT '交易个数',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_hash` (`hash`),
    KEY `idx_height`(`height`),
    KEY `idx_preblock`(`previousblockhash`),
    KEY `idx_time`(`time`)   
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS btc_database.t_transaction_info (
    `id`                BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `blockhash`         CHAR(64)            NOT NULL DEFAULT ''     COMMENT '当前交易所在区块的哈希',
    `blockheight`       INT                 NOT NULL DEFAULT 0      COMMENT '区块高度',
    `time`              BIGINT              NOT NULL DEFAULT 0      COMMENT '区块打包时间',
    `isfork`            TINYINT             NOT NULL DEFAULT 0      COMMENT '是否为分叉链，0表示为主链，1表示为分叉链',
    `txid`              CHAR(64)            NOT NULL DEFAULT ''     COMMENT '交易哈希',
    `iscoinbase`        TINYINT             NOT NULL DEFAULT 0      COMMENT '是否为coinbase，0表示普通交易，1表示为coinbase',
    `hash`              CHAR(64)            NOT NULL DEFAULT ''     COMMENT 'Witness哈希',
    `size`              INT                 NOT NULL DEFAULT 0      COMMENT '交易大小',
    `vsize`             BIGINT              NOT NULL DEFAULT 0      COMMENT '加权大小',
    `version`           INT                 NOT NULL DEFAULT 0      COMMENT '版本号',
    `locktime`          BIGINT              NOT NULL DEFAULT 0      COMMENT '锁定时间',
    `weight`            BIGINT              NOT NULL DEFAULT 0      COMMENT '权重',
   
    PRIMARY KEY (`id`),
    KEY `idx_txid` (`txid`),
    KEY `idx_blockhash`(`blockhash`) ,
    KEY `idx_time`(`time`),
    UNIQUE KEY `uniq_blockhash_txid`(`blockhash`, `txid`),
    FOREIGN KEY `fr_blockhash`(blockhash) REFERENCES t_block_info(hash) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS btc_database.t_input_info (
    `id`                BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `blockhash`         CHAR(64)            NOT NULL DEFAULT ''     COMMENT '当前交易所在区块的哈希',
    `time`              BIGINT              NOT NULL DEFAULT 0      COMMENT '区块打包时间',
    `isfork`            TINYINT             NOT NULL DEFAULT 0      COMMENT '是否为分叉链，0表示为主链，1表示为分叉链',
    `hash`              CHAR(64)            NOT NULL DEFAULT ''     COMMENT '所在交易哈希',
    `txid`              CHAR(64)            NOT NULL DEFAULT ''     COMMENT '交易id',
    `vout`              BIGINT              NOT NULL DEFAULT 0      COMMENT 'vout索引号',
    `sequence`          BIGINT              NOT NULL DEFAULT 0      COMMENT '脚本序列号',
    `hex`               TEXT				NOT NULL				COMMENT '交易的签名信息脚本的十六进制表示',
    `asm`               TEXT				NOT NULL				COMMENT '交易的签名信息脚本的asm码表示',
    `coinbase`          TEXT				NOT NULL				COMMENT 'coinbase的scriptSig信息',
    `from`              VARCHAR(64)         NOT NULL DEFAULT ''     COMMENT '转出账户地址',
    `value`             BIGINT              NOT NULL DEFAULT 0      COMMENT '转账金额',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`hash`),
    KEY `idx_txid`(`txid`),
    KEY `idx_from`(`from`),
    key `idx_from_time`(`from`, `time`),
    key `idx_txid_vout`(`txid`, `vout`),
    UNIQUE KEY `uniq_blockhash_hash_txid_vout`(`blockhash`, `hash`, `txid`, `vout`),
    FOREIGN KEY `fr_input_hash`(hash) REFERENCES t_transaction_info(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS btc_database.t_output_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `blockhash`         CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '当前交易所在区块的哈希',
    `time`              BIGINT                              NOT NULL DEFAULT 0          COMMENT '区块打包时间',
    `isfork`            TINYINT                             NOT NULL DEFAULT 0          COMMENT '是否为分叉链，0表示为主链，1表示为分叉链',
    `hash`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '所在交易哈希',
    `value`             BIGINT                              NOT NULL DEFAULT 0          COMMENT '转账金额',
    `n`                 BIGINT                              NOT NULL DEFAULT 0          COMMENT '索引号',
    `hex`               TEXT								NOT NULL					COMMENT '公钥哈希脚本的十六进制表示',
    `asm`               TEXT								NOT NULL					COMMENT '公钥哈希脚本的asm表示',
    `type`              varchar(64)                         NOT NULL DEFAULT ''         COMMENT '公钥哈希脚本的类型',
    `reqSigs`           INT                                 NOT NULL DEFAULT 0          COMMENT '需要签名的个数',
    `to`                varchar(64)                         NOT NULL DEFAULT ''         COMMENT '转入账户地址',
    `state`             TINYINT                             NOT NULL DEFAULT 0          COMMENT '当前花费状态，0表示未花费，1表示已花费，2表示被未确认交易花费',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`hash`),
    KEY `idx_to`(`to`),
    key `idx_to_time`(`to`, `time`),
    key `idx_to_state`(`to`, `state`),
    key `idx_hash_n`(`hash`, `n`),
    UNIQUE KEY `uniq_blockhash_hash_n`(`blockhash`,`hash`, `n`),
    FOREIGN KEY `fr_output_hash`(hash) REFERENCES t_transaction_info(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS btc_database.t_output_address_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `blockhash`         CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '当前交易所在区块的哈希',
    `hash`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '所在交易哈希',
    `n`                 BIGINT                              NOT NULL DEFAULT 0          COMMENT '索引号',
    `address`           varchar(64)							NOT NULL DEFAULT ''			COMMENT 'output的地址信息',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`hash`),
    KEY `idx_blockhash`(`blockhash`),
    KEY `idx_address`(`address`),
    FOREIGN KEY `fr_output_address_hash`(hash) REFERENCES `t_output_info`(hash) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

# 交易记录中间表
CREATE TABLE IF NOT EXISTS btc_database.t_transaction_input_output_address_info (
    `id`                BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `blockhash`         CHAR(64)            NOT NULL DEFAULT ''         COMMENT '当前交易所在区块的哈希',
    `time`              BIGINT              NOT NULL DEFAULT 0          COMMENT '区块打包时间',
    `txid`              CHAR(64)            NOT NULL DEFAULT ''         COMMENT '所在交易哈希',
    `address`           VARCHAR(64)         NOT NULL DEFAULT ''         COMMENT '转出账户地址',
    `isfrom`            TINYINT             NOT NULL DEFAULT 0          COMMENT '是否为input，0表示output，1表示为input',

    PRIMARY KEY (`id`),
    KEY `idx_txid`(`txid`),
    KEY `idx_address`(`address`),
    key `idx_address_time`(`address`, `time`),
    UNIQUE KEY `uniq_address_index_isfrom`(`txid`, `address`, `isfrom`)
    FOREIGN KEY `fr_transaction_hash`(txid) REFERENCES t_transaction_info(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;


# create VIEW
CREATE VIEW btc_database.v_transaction_info(blockhash, blockheight, time, transactionhash, iscoinbase, fromhash, fromindex, fromvalue, fromaddress, coinbase, toindex, tovalue, toaddress, totype, toasm, state) 
AS SELECT t1.blockhash, t1.blockheight, t1.time, t1.txid, t1.iscoinbase, 
t2.txid, t2.vout, t2.value, t2.from, t2.coinbase, 
t3.n, t3.value, t3.to, t3.type, t3.asm, t3.state FROM btc_database.t_transaction_info t1 
INNER JOIN btc_database.t_input_info t2 ON t1.txid=t2.hash 
INNER JOIN btc_database.t_output_info t3 ON t1.txid=t3.hash;

# omni transaction
CREATE TABLE IF NOT EXISTS btc_database.t_omni_transaction_info (
    `id`                            BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `txid`                          CHAR(64)            NOT NULL DEFAULT ''         COMMENT '交易哈希',
    `blockhash`                     CHAR(64)            NOT NULL DEFAULT ''         COMMENT '当前交易所在区块的哈希',
    `blocktime`                     BIGINT              NOT NULL DEFAULT 0          COMMENT '区块打包时间',
    `block`                         INT                 NOT NULL DEFAULT 0          COMMENT '区块高度',
    `positioninblock`               INT                 NOT NULL DEFAULT 0          COMMENT '在区块中的位置',
    `fee`                           CHAR(64)            NOT NULL DEFAULT ''         COMMENT '手续费',
    `sendingaddress`                VARCHAR(64)         NOT NULL DEFAULT ''         COMMENT '转出账户地址',
    `referenceaddress`              varchar(64)         NOT NULL DEFAULT ''         COMMENT '转入账户地址',
    `type_int`                      BIGINT              NOT NULL DEFAULT 0          COMMENT '交易类型',
    `type`                          varchar(64)         NOT NULL DEFAULT ''         COMMENT '交易类型',
    `propertyid`                    BIGINT              NOT NULL DEFAULT 0          COMMENT '财产编号',
    `amount`                        varchar(64)         NOT NULL DEFAULT ''         COMMENT '代币转账金额',
    `ismine`                        TINYINT             NOT NULL DEFAULT 0          COMMENT '是否为转给自己，0表示不是转给自己，1表示是转给自己',
    `version`                       BIGINT              NOT NULL DEFAULT 0          COMMENT '版本号',
    `valid`                         TINYINT             NOT NULL DEFAULT 0          COMMENT '是否为有效的，0表示不是有效的，1表示是有效的',
    `invalidreason`                 TEXT				NOT NULL					COMMENT '无效的原因',
    `divisible`                     TINYINT             NOT NULL DEFAULT 0          COMMENT '是否为可分割的，0表示不是可分割的，1表示是可分割的',
    `purchasedpropertyid`           BIGINT              NOT NULL DEFAULT 0          COMMENT '购买的财产id',
    `purchasedpropertyname`         varchar(64)         NOT NULL DEFAULT ''         COMMENT '购买的财产的名称',
    `purchasedpropertydivisible`    TINYINT             NOT NULL DEFAULT 0          COMMENT '财产是否可分割，0表示不可分割，1表示可分割',
    `purchasedtokens`               varchar(64)         NOT NULL DEFAULT ''         COMMENT '购买物品tokens',
    `issuertokens`                  varchar(64)         NOT NULL DEFAULT ''         COMMENT '发行人tokens',
	`totalstofee`                   varchar(64)         NOT NULL DEFAULT ''         COMMENT 'omni费用',
	`ecosystem`                     varchar(64)         NOT NULL DEFAULT ''         COMMENT '系统类型',
	`bitcoindesired`                varchar(64)         NOT NULL DEFAULT ''         COMMENT '渴望的数量',
	`timelimit`                     INT                 NOT NULL DEFAULT 0          COMMENT '时限',
	`feerequired`                   varchar(64)         NOT NULL DEFAULT ''         COMMENT '需要的花费',
	`action`                        varchar(64)         NOT NULL DEFAULT ''         COMMENT '操作',
	`propertyidforsale`             BIGINT              NOT NULL DEFAULT 0          COMMENT '卖出的财产id',
	`propertyidforsaleisdivisible`  TINYINT             NOT NULL DEFAULT 0          COMMENT '卖出财产是否可分割，0表示不可分割，1表示可分割',
	`amountforsale`                 varchar(64)         NOT NULL DEFAULT ''         COMMENT '要卖出的数量',
	`propertyiddesired`             BIGINT              NOT NULL DEFAULT 0          COMMENT '想要买入的财产id',
	`propertyiddesiredisdivisible`  TINYINT             NOT NULL DEFAULT 0          COMMENT '想要买入的财产是否可分割，0表示不可分割，1表示可分割',
	`amountdesired`                 varchar(64)         NOT NULL DEFAULT ''         COMMENT '要买入的财产的数量',
	`unitprice`                     varchar(64)         NOT NULL DEFAULT ''         COMMENT '单价',
	`amountremaining`               varchar(64)         NOT NULL DEFAULT ''         COMMENT '保留的数量',
	`amounttofill`                  varchar(64)         NOT NULL DEFAULT ''         COMMENT '填充的数量',
	`status`                        varchar(64)         NOT NULL DEFAULT ''         COMMENT '状态',
	`canceltxid`                    varchar(64)         NOT NULL DEFAULT ''         COMMENT '取消的交易id',
	`propertytype`                  varchar(64)         NOT NULL DEFAULT ''         COMMENT '财产类型',
	`category`                      varchar(512)        NOT NULL DEFAULT ''         COMMENT '类别',
	`subcategory`                   varchar(512)        NOT NULL DEFAULT ''         COMMENT '子类别',
	`propertyname`                  varchar(512)         NOT NULL DEFAULT ''        COMMENT '财产名称',
	`data`                          varchar(512)        NOT NULL DEFAULT ''         COMMENT '数据',
	`url`                           varchar(512)        NOT NULL DEFAULT ''         COMMENT 'url',
	`tokensperunit`                 varchar(64)         NOT NULL DEFAULT ''         COMMENT 'token数量',
	`deadline`                      BIGINT              NOT NULL DEFAULT 0          COMMENT '最后期限',
	`earlybonus`                    INT                 NOT NULL DEFAULT 0          COMMENT '早鸟红利',
	`percenttoissuer`               INT                 NOT NULL DEFAULT 0          COMMENT '发行百分比',
	`featureid`                     BIGINT              NOT NULL DEFAULT 0          COMMENT '特性id',
	`activationblock`               BIGINT              NOT NULL DEFAULT 0          COMMENT '激活区块高度',
	`minimumversion`                BIGINT              NOT NULL DEFAULT 0          COMMENT '最低版本',

    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_txid` (`txid`),
    KEY `idx_blockhash`(`blockhash`) ,
    KEY `idx_time`(`blocktime`),
    KEY `idx_height`(`block`),
    KEY `idx_send`(`sendingaddress`),
    KEY `idx_reference`(`referenceaddress`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS btc_database.t_omni_purchase_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `txid`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '交易哈希',
    `vout`              BIGINT                              NOT NULL DEFAULT 0          COMMENT 'vout',
	`amountpaid`        varchar(64)							NOT NULL DEFAULT ''			COMMENT '支出金额',
	`ismine`            TINYINT                             NOT NULL DEFAULT 0          COMMENT '是否为转给自己，0表示不是转给自己，1表示是转给自己',
	`referenceaddress`  varchar(64)                         NOT NULL DEFAULT ''         COMMENT '转入账户地址',
	`propertyid`        BIGINT                              NOT NULL DEFAULT 0          COMMENT '财产编号',
	`amountbought`      varchar(64)                         NOT NULL DEFAULT ''         COMMENT '购买数量',
	`valid`             TINYINT                             NOT NULL DEFAULT 0          COMMENT '是否为有效的，0表示不是有效的，1表示是有效的',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`txid`),
    KEY `idx_address`(`referenceaddress`),
    FOREIGN KEY `fr_purchase_omni_transaction_hash`(txid) REFERENCES `t_omni_transaction_info`(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS btc_database.t_omni_recipient_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `txid`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '交易哈希',
    `address`           varchar(64)                         NOT NULL DEFAULT ''         COMMENT '地址',
	`amount`            varchar(64)                         NOT NULL DEFAULT ''         COMMENT '数量',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`txid`),
    FOREIGN KEY `fr_recipient_omni_transaction_hash`(txid) REFERENCES `t_omni_transaction_info`(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS btc_database.t_omni_subsend_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `txid`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '交易哈希',
    `propertyid`        BIGINT                              NOT NULL DEFAULT 0          COMMENT '财产编号',
	`divisible`         TINYINT                             NOT NULL DEFAULT 0          COMMENT '是否为可分割的，0表示不是可分割的，1表示是可分割的',
	`amount`            varchar(64)                         NOT NULL DEFAULT ''         COMMENT '数量',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`txid`),
    FOREIGN KEY `fr_subsend_omni_transaction_hash`(txid) REFERENCES `t_omni_transaction_info`(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS btc_database.t_omni_trade_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `hash`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '所属交易哈希',
    `txid`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '交易哈希',
    `block`             INT                                 NOT NULL DEFAULT 0          COMMENT '区块高度',
	`address`           varchar(64)                         NOT NULL DEFAULT ''         COMMENT '地址',
	`amountsold`        varchar(64)                         NOT NULL DEFAULT ''         COMMENT '销售数量',
	`amountreceived`    varchar(64)                         NOT NULL DEFAULT ''         COMMENT '接收数量',
	`tradingfee`        varchar(64)                         NOT NULL DEFAULT ''         COMMENT '交易费用',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`txid`),
    KEY `idx_address`(`address`),
    FOREIGN KEY `fr_trade_omni_transaction_hash`(hash) REFERENCES `t_omni_transaction_info`(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS btc_database.t_omni_cancel_info (
    `id`                BIGINT UNSIGNED                     NOT NULL AUTO_INCREMENT     COMMENT '自增主键',
    `hash`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '所属交易哈希',
    `txid`              CHAR(64)                            NOT NULL DEFAULT ''         COMMENT '交易哈希',
    `propertyid`        BIGINT                              NOT NULL DEFAULT 0          COMMENT '财产编号',
	`amountunreserved`  varchar(64)                         NOT NULL DEFAULT ''         COMMENT '释放的数量',

    PRIMARY KEY (`id`),
    KEY `idx_hash`(`txid`),
    FOREIGN KEY `fr_cancel_omni_transaction_hash`(hash) REFERENCES `t_omni_transaction_info`(txid) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4;

# create user btc
DROP USER 'btc'@'%';
CREATE USER 'btc'@'%' IDENTIFIED BY '#Bitcoin_2019';

# grant user btc
GRANT ALL ON btc_database.* TO 'btc'@'%';
