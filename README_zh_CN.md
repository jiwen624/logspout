# LogSpout: 一个简单的日志生成工具

## 简介
LogSpout可根据用户提供的样本日志, 通过正则表达式配置来替换特定位置的文本, 并指定产生速率, 最终生成新的日志流.

目前支持的特性:

1. 控制日志产生速度(毫秒级), 也可设置hightide=true使用压测模式(每条日志中间没有间隔时间).
2. 产生的日志输出到stdout, 用户可自行重定向到文件或其他输出端.
3. 支持通过外部文件配置样本日志以及替换值列表.
4. 支持通过正则表达式指定要做替换的字段.
5. 支持对每个字段单独配置替换规则.
6. 目前支持的替换规则: 时间戳(timestamp), 固定列表(fixed-list), 数字(integer), IPv4/v6地址(区分国内外), 国内手机号码等.
7. 支持以随机/递增/递减方式获取替换字段值. 随机方式目前用近似高斯分布, 更贴近实际数据分布情况.
8. 增加looks-real数据替换选项, 可生成IPv4/IPv6地址, email地址, 人名, 国家, 浏览器User Agent等.

## 使用方式
logspout默认使用logspout.json做为配置文件, 如果该文件存在且配置合法, 则直接运行:

```./logspout```

就会输出生成的日志到stdout.

如果想将日志重定向到某个文件, 则可以使用:

```./logspout > my.log```

可以使用-h选项获取命令帮助:
```
➜  logspout git:(master) ✗ ./logspout -h
Usage of ./logspout:
  -f string
    	specify the config file in json format. (default "logspout.json")
  -v string
    	Print level: debug, info, warning, error. (default "warning")
```

简单来说, 可以用-f选项指定你的配置文件, 以及使用-v debug/info/warning/error指定该工具自己的日志打印级别.

如果你发现程序运行和预想的不同, 可以开启debug模式(`-v debug`), 打印详细日志进行定位.

**注意**: 为了与生成的机器日志区分, logspout自己的日志默认是全部输出到stderr的.


## 配置说明

为灵活性考虑, 配置文件以标准json格式提供, 另随代码附带了一个[示例配置文件](https://github.com/jiwen624/logspout/blob/master/logspout.json). 因为是json所以配置文件不支持注释, 这是一个天然的缺陷.

配置项如下:
### hightide
**说明**:

默认为false, 表示每条消息之间会有一段思考时间(可通过下面的min-interval和max-interval配置).
为true则开启压测模式, 每条日志之间没有思考时间, 配合concurrency设置可以占满服务器的CPU, I/O, 网卡带宽等, 谨慎使用.

### concurrency
**说明**:

配置生成日志的并发数, 默认为1, 可配置>1的数字, 以便增加日志产生速率. 具体的并发度需要根据服务器硬件情况调整.

一个参考例子: 在我的2014 Macbook Pro上, 配置concurrency=1000, min-interval=100, max-interval=500, 产生日志速率约为10,0000条/min,
注意此数字受限于max/min-interval, 因此时CPU占用只有约15%.



### min-interval
**说明**: 产生下一条新日志的最小时间间隔

**单位**: ms(毫秒)

### max-interval
**说明**: 产生下一条新日志的最大时间间隔

**单位**: ms(毫秒)

**Tips**: 如果想配置规整的日志发送间隔, 可将min-interval和max-interval设置为相同的值, 否则logspout每次产生新日志前会在min/max-interval之
间随机选取一个值做为间隔.


### logtype
**说明**:

暂时没实质性用处, 可以做为日志类型的标记.

### file
**说明**:

样本日志所在的文件, 只放一条样本日志在此文件内. 如果有多条, logspout也会一次性全部读入内存, 并进行正则匹配.

### pattern
**说明**:

从样本日志抽取字段的正则表达式.

**注意**:

正则表达式源远流长, 流派众多, 即使是PCRE之类用途广泛的, 各个语言的支持也有些不同. logspout正则使用的是Go/Python的格式(就目前所见).
但是有在考虑做一层预处理, 减轻用户工作量.

目前此工具要求raw message里所有的文本均配置成captured group, 即使你不需要替换它. 对于不需要替换的部分, 可以只用()包围起来, 不用起名字.


**示例**:

(注意捕获的字段用(?P<name>)而不是(?<name>), 另外`\`需要用双斜杠取消转义`\\`)

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
此时需要定义`"min"`和`"max"`, 提供一个选择范围.

### looks-real
```"type": "looks-real"```

生成仿真的特性类型数据. 可以使用method指定要生成哪种类型的数据, 目前支持:

`"method": "ipv4"`  - IPv4 地址

`"method": "ipv4china"`  - 中国IPv4地址

`"method": "cellphone-china"`  - 中国手机号码

`"method": "ipv6"`  - IPv6 地址

`"method": "country"`  - 国家名称

`"method": "email"`  - email地址

`"method": "name"`  - 人名

`"method": "user-agent"`  - 浏览器的User Agent信息


## 疑问
如有疑问或发现可提交issues.



