# LogSpout: 一个简单的日志生成工具

## 简介
LogSpout可根据用户提供的样本日志, 通过正则表达式配置来替换特定位置的文本, 并指定产生速率, 最终生成新的日志流.

目前支持的特性:

1. 控制日志产生速度(毫秒级), 也可设置`"hightide"=true`使用压测模式(每条日志中间没有间隔时间).
2. 产生的日志输出到`stdout`, 用户可自行重定向到文件或其他输出端.
3. 支持通过外部文件配置样本日志以及替换值列表. 支持多行日志.
4. 支持通过正则表达式指定要做替换的字段.
5. 支持对每个字段单独配置替换规则.
6. 目前支持的替换规则: 时间戳(`timestamp`), 固定列表(`fixed-list`), 数字(`integer`), IPv4/v6地址(区分国内外), 国内手机号码等.
7. 支持以随机/递增/递减方式获取替换字段值. 随机方式目前用近似高斯分布, 更贴近实际数据分布情况.
8. 增加looks-real数据替换选项, 可生成IPv4/IPv6地址, email地址, 人名, 国家, 浏览器User Agent等.
9. 支持事务模式(配置多条events组成一个事务)
10. 支持多线程打印日志, 可通过`concurrency`配置并发线程数.
11. 支持以udp/tcp方式输出到syslog (目前日志级别暂时固定为INFO)
12. 支持同时写入多个日志文件(方便压测性能), 通过`duplicate`参数指定, 文件名加入N_前缀
13. 通过web console获取最近1s的EPS.
14. 支持随机生成XML和JSON文档, 且其maximum depth和maximum elements, 以及tag-seed可配置.
15. 支持限定最多可以产生的日志条数, 通过`max-events`参数控制.
16. 支持查看当前运行的配置文件名称及配置详情.

## 使用方式

logspout默认使用logspout.json做为配置文件, 如果该文件存在且配置合法, 则直接运行:

```./logspout```

就会输出生成的日志到stdout.

如果想将日志重定向到某个文件, 则可以使用输出重定向:

```./logspout > my.log```

或者使用```output```参数(参见下文的配置介绍)

LogSpout提供了一些访问接口, 可以查看当前的运行情况:

```
curl http://your-host:10306/counter?details=true 获取各个worker的当前的eps.
curl http://your-host:10306/counter 获取当前总的eps.
curl http://your-host:10306/config  获取当前运行的配置文件名
curl http://your-host:10306/config?details=true 获取当前运行的配置
```

其中10306为默认端口, 如果有冲突, 请修改配置文件中的`console-port`参数.

如果不需要看每个worker的eps, 可以去掉`?details=true`

可以使用-h选项获取命令帮助:
```
➜  logspout git:(master) ✗ ./logspout -h
Usage of ./logspout:
  -f string
    	specify the config file in json format. (default "logspout.json")
  -v string
    	Print level: debug, info, warning, error. (default "info")
```

简单来说, 可以用-f选项指定你的配置文件, 以及使用-v debug/info/warning/error指定该工具自己的日志打印级别.

如果你发现程序运行和预想的不同, 可以开启debug模式(`-v debug`), 打印详细日志进行定位.

**注意**: 为了与生成的机器日志区分, logspout自己的日志默认是全部输出到stderr的.


## 配置说明

为灵活性考虑, 配置文件以标准json格式提供(因为是json所以配置文件不支持注释, 这是一个天然的缺陷, 以后计划改成yaml格式). 

下面是一个完整的配置文件示例, 该文件也可在此处找到: [示例配置文件](https://github.com/jiwen624/logspout/blob/master/logspout.json).

```
{
	"hightide": false,
	"concurrency": 100,
	"min-interval": 100,
	"max-interval": 1010,
	"duration": 3600,
	"logtype": "weblogic",
	"sample-file": "samples/sample.log",
	"output-stdout": true,
	"output-syslog": {
		"protocol": "udp",
		"netaddr": "localhost:514",
		"tag": "logspout"
	},
	"output-file": {
		"file-name": "default.log",
		"max-size": 5,
		"max-backups": 3,
		"compress": false,
		"max-age":7
	},
	"pattern": "(####<)(?<timestamp>.*?)(>\s*<)(?<severity>.*?)(>\s*<)(?<subsystem>.*?)(>\s*<)(?<ipaddress>.*?)(>\s*<)(?<phone>.*?)(>\s*<)(?<thread>.*?)(>\s*<)(?<user>.*?)(>\s*<)(?<transaction>.*?)(>\s*<)(?<diagcontext>.*?)(>\s*<)(?<rawtime>.*?)(>\s*<BEA-)(?<msgid>.*?)(>\s*<)(?<msgtext>.*?)(>)",
	"replacement": {
		"timestamp": {
			"type": "timestamp",
			"format": "MMM dd, yyyy hh:mm:ss.SSS a z"
		},
		"severity": {
			"type": "fixed-list",
			"method": "random",
			"list-file": "samples/severity.sample"
		},
		"subsystem": {
			"type": "float",
			"min": 100,
			"max": 10000,
			"precision": 20
		},
		"ipaddress": {
			"type": "looks-real",
			"method": "ipv4china"
		},
		"phone": {
			"type": "looks-real",
			"method": "cellphone-china"
		},
		"user": {
			"type": "fixed-list",
			"method": "random",
			"list": ["GuoJing", "HuangRong", "ZhangSanfeng", "ZhangWuji","LiMochou", "OuyangFeng", "HongQigong", "LinghuChong", "RenYingying", "DuanTiande"]
		},
		"thread": {
			"type": "integer",
			"method": "prev",
			"min": 0,
			"max": 10
		},
		"transaction": {
			"type": "string",
			"chars": "abcde12345",
			"min": 10,
			"max": 20
		},
		"msgid": {
			"type": "integer",
			"method": "random",
			"min": 0,
			"max": 2000000
		}
	}

}
```


**配置项详细说明如下:**
### hightide
**说明**:

默认为false, 表示每条消息之间会有一段思考时间(可通过下面的min-interval和max-interval配置).
为true则开启压测模式, 每条日志之间没有思考时间, 配合concurrency设置可以占满服务器的CPU, I/O, 网卡带宽等, 谨慎使用.

### uniform
**说明**

默认为true, 表示压力是均匀分布的. 此时采用有尾部截断的高斯分布产生随机数做为随机思考时间.

如果设置为false, 则模拟大多数场景下压力随着时间波动的情景, 在上午和下午各有一个波峰, 晚上则是一天之中压力最小的时候. 峰值与波谷的压力差
大约为5-6倍左右.

目前没有加入在分钟或者小时粒度上的随机波动.

### concurrency
**说明**:

配置生成日志的并发数, 默认为1, 可配置>1的数字, 以便增加日志产生速率. 具体的并发度需要根据服务器硬件情况调整.

一个参考例子: 在我的2014 Macbook Pro上, 配置concurrency=1000, min-interval=100, max-interval=500, 产生日志速率约为10,0000条/min,
注意此数字受限于max/min-interval, 因此时CPU占用只有约15%.

如果单纯想提高eps, 可以通过如下方式进行:

1. 减少`min-interval`和`max-interval`增加日志发送频次

2. 使用`hightide=true`模式, 关闭每条日志之间的think time

3. 通过配置较大的`concurrency`来增加worker的数量

4. 通过`duplicate`参数将日志同时写入多个目标文件

5. 适当减少字段替换, 也可以大幅提升eps 


### duplicate
**说明**:

设置同时写入的文件数, 默认为1. 通过此参数可以讲同样的一条日志写入多个日志文件, 相当于将日志eps数值放大N倍.

### console-port
**说明**:

设置此logspout进程的web console的端口, 默认为10306, 如果同时启动多个logspout, 则需要修改此处为其他的端口, 避免互相冲突

### min-interval
**说明**: 产生下一条新日志的最小时间间隔

**单位**: ms(毫秒)

### max-interval
**说明**: 产生下一条新日志的最大时间间隔

**单位**: ms(毫秒)

**Tips**: 如果想配置规整的日志发送间隔, 可将min-interval和max-interval设置为相同的值, 否则logspout每次产生新日志前会在min/max-interval之
间随机选取一个值做为间隔.


### duration
**说明**: 程序运行时长

**单位**: s(秒)

**Tips**: 默认值为0, 表示无限期运行, 永不停止.


### max-events
**说明**: 程序产生的最大日志条数, 超过这个数字则退出.

**单位**: 无

**Tips**: 默认值为最大的uint64值

### logtype
**说明**:

暂时没实质性用处, 可以做为日志类型的标记.

### sample-file
**说明**:

样本日志所在的文件, 在没有配置事务模式(默认没有开启事务模式, 事务模式指的是由多条样本日志组成一个完整事务, 每条日志各自有正则匹配).

只放一条样本日志在此文件内. 如果有多条, logspout也会一次性全部读入内存, 并当做一条日志进行正则匹配. (此处暂时没有优化, 因此请不要
放太多日志在sample文件, 一条即可.)

### output-syslog
**说明**:

日志输出到syslog, 目前级别固定为INFO, 后续可能会增加其他级别的配置项.
```
"output-syslog": {
    "protocol": "udp",   (可配置为udp或tcp)
    "netaddr": "localhost:514",   (syslog接收地址)
    "tag": "logspout"      (标签)
}
```

### output-file
**说明**:

日志输出文件.

可配置如下几个参数, 控制输出文件的滚动, 最大文件大小, 最大保留天数等:
```
"output-file": {
    "file-name": "default.log",
    "max-size": 5,      (单位: MB)
    "max-backups": 3,   (如果配置为3, 则加上当前日志文件一共最多保留4个文件, 之后最老的文件会被删除掉)
    "compress": false,  (默认对备份的文件不配置压缩)
    "max-age":7         (最大保留天数)
 }
```

### 事务(transaction)支持
使用如下配置.

```
 15         "transaction": true,
 16         "transaction-ids": ["thread", "msgid"],
 17         "max-intra-transaction-latency": 40,
 18         "pattern": ["(Start####<)(?<timestamp>.*?)(>\s*<)(?<severity>.*?)(>\s*<)(?<subsystem>.*?)(>\s*<)(?<ipaddress>.*?)(>\s*<)(?<phone>    .*?)(>\s*<)(?<thread>.*?)(>\s*<)(?<user>.*?)(>\s*<)(?<transaction>.*?)(>\s*<)(?<diagcontext>.*?)(>\s*<)(?<rawtime>.*?)(>\s*<BEA-)(?<msgid    >.*?)(>\s*<)(?<msgtext>.*?)(>)",
 19         "(End####<)(?<timestamp>.*?)(>\s*<)(?<severity>.*?)(>\s*<)(?<subsystem>.*?)(>\s*<)(?<ipaddress>.*?)(>\s*<)(?<phone>.*?)(>\s*<)(?<    thread>.*?)(>\s*<)(?<user>.*?)(>\s*<)(?<transaction>.*?)(>\s*<)(?<diagcontext>.*?)(>\s*<)(?<rawtime>.*?)(>\s*<BEA-)(?<msgid>.*?)(>\s*<)(?    <msgtext>.*?)(>)"],
```

注意:
1. sample log文件里的一笔事务的多条日志要使用空行分开, 否则会被认为是一条多行日志.

2. pattern使用数组方式按顺序指定每条日志的正则匹配方式. (非事务模式下可以直接使用字符串)

3. `transaction_ids`配置事务的关联id (一笔事务的各条日志中, 这些字段值是一样的)

4. `max-intra-transaction-latency` 配置一笔事务中各条日志之间的最大时间间隔, 实际间隔会随机在该范围内选一个数值

5. 使用事务模式, 需要把`transaction`设置为`true`.

### pattern
**说明**:

用此参数定义正则表达式, 用于从样本日志抽取字段.

正则表达式源远流长, 流派众多, 即使是PCRE之类用途广泛的, 各个语言的支持也有些不同. logspout正则使用的是Go/Python的格式(就目前所见).
但是做了一层预处理, 减轻用户工作量.

**注意**:

1. 目前此工具要求raw message里所有的文本均配置成captured group, 即使你不需要替换它. 对于不需要替换的部分, 可以只用()包围起来, 不用起名字.**

(在以后的版本中会通过预处理自动化这部分工作, 届时用户可直接将字段解析的正则表达式拷贝过来即可)

**示例**:

(下面示例中捕captured group用(?P<name>)而不是(?<name>), 这是Python/Perl/Go的re语法, 不过也可以写成(?<name>), logspout内部会做预处理.)

```
  "pattern": "(####<)(?P<timestamp>.*?)(>\\s*<)(?P<severity>.*?)(>\\s*<)(?P<subsystem>.*?)(>\\s*<)(?P<machine>.*?)(>\\s*<)(?P<serve    r>.*?)(>\\s*<)(?P<thread>.*?)(>\\s*<)(?P<user>.*?)(>\\s*<)(?P<transaction>.*?)(>\\s*<)(?P<diagcontext>.*?)(>\\s*<)(?P<rawtime>.*?)(>\\s*<    BEA-)(?P<msgid>.*?)(>\\s*<)(?P<msgtext>.*?)(>)"
```

### replacement
**说明**:

定义正则表达式里捕获的每个字段的替换规则.

replacement内的每个key都是pattern里的一个captured group, 通过此处配置替换规则.
具体可以参考样板配置文件logspout.json

每个字段里, 使用type定义替换规则, 对每个不同的type, 有不同的其他字段要求.
目前支持的替换规则简述如下:

### timestamp
```"type": "timestamp"```
时间戳, 此时需要定义format, 指定时间戳的格式(支持标准的joda格式时间戳)

### fixed-list

```"type": "fixed-list"```
从指定的固定列表里选择该字段的替换值.
该固定列表可以直接用列表指定:
```"list": ["aaa", "bbb"]```

也可以从外部文件读取(如果该列表较大. 使用外部文件时, 文件的每一行做为一个备选值)
```"list-file": "/path/to/file"```

此外可以定义从列表里选取值的方式: 随机/递增/递减
```"method": "random"```  (或者`next`, `prev`)

### integer
```"type": "integer"```

用数字做为该字段的替换值.

可以定义从列表里选取值的方式: 随机/递增/递减

```"method": "random"```  (或者`next`, `prev`)

注意: 如果想通过`next`, `min`, `max`三个参数生成一个自增ID序列, 请设置`concurrency`参数为1, 目前暂时不支持多个worker共享一个全局的自增ID.

此时需要定义`"min"`和`"max"`, 提供一个选择范围.

### float
```"type": "float"```

用浮点数做为该字段的替换值. 此类型只能使用随机值.
可配置的选项:
```"min": 100,
   "max": 1000,
   "precision": 10,
```
precision表示浮点数精度, 小数点后保留几位数字.

### string
```"type": "string"```

用字符串做为该字段的替换值. 此类型只能使用随机值, 可定义字符串的最大最小长度.
可配置的选项:
```"min": 10,
   "max": 20,
```

你也可以用`"chars": "abcd1234"`参数指定随机字符串的字符选择范围, 默认不需要指定此参数, 字符选择范围为[a-zA-Z]

### looks-real
```"type": "looks-real"```

生成仿真的特性类型数据. 可以使用method指定要生成哪种类型的数据, 目前支持:

`"method": "ipv4"`  - IPv4 地址

`"method": "ipv4china"`  - 中国IPv4地址

`"method": "cellphone-china"`  - 中国手机号码

`"method": "ipv6"`  - IPv6 地址

`"method": "mac"`  - Mac 地址

`"method": "country"`  - 国家名称

`"method": "email"`  - email地址

`"method": "name"`  - 人名

`"method": "chinese-name"`  - 中国人名

`"method": "user-agent"`  - 浏览器的User Agent信息

`"method": "uuid"`  - UUID

`"method": "xml"`   - 随机生成XML文档

`"method: "json"`   - 随机生成JSON文档

其中`xml`和`json`支持配置最大嵌套深度, 每个层次的最大元素个数以及tag names的种子数据(如果不指定`tag-seed`, 则使用默认的300+ distinct values的种子数据):

```
"msgtext": {
        "type": "looks-real",
        "method": "xml"
        "parms": {
               "max-depth":10,
               "max-elements":100,
               "tag-seed": ["one", "two", "three"]
},
```

增加这三项配置, 主要是因为在xml日志采集到日志分析引擎(如ElasticSearch)并解析字段的时候, 字段个数一般不宜过大(ES mappings, etc.)
因此通过xml文档的深度, 每层的最大元素数, 以及备选的tag名字进行限制. 

## 疑问/Bugs
如有疑问或发现Bug可提交issues, 并附上问题出现时的样本日志和logspout.json配置.



