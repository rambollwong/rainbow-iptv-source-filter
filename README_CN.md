# Rainbow IPTV Source Filter

![](https://ramboll.wang/img/IPTVSourceFilter_LOGO_SD.png)

[[English]](./README.md)

`Rainbow IPTV Source Filter` 是一个用于过滤 IPTV 源的工具。它能够检测并过滤掉不可达或质量较差的源，并将可用的源合并到指定的输出文件中。

## ⚠️ 注意事项

1. 本工具适用于在与 IPTV 播放设备相同的网络环境下，对直播源进行本地可用性测试，并将过滤后的有效源合并到目标文件中。请勿在云服务器或与播放设备网络环境不一致的场景下使用本工具。
2. 本工具默认使用 `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36` 作为测试访问的 User-Agent（UA）。如需修改，请在配置文件中设置自定义 UA。若使用默认 UA，播放过滤后的源时，可能需要调整播放器的 UA 设置以确保兼容性（例如：天光云影播放器）。

## ⚠️ 免责声明

1. 本工具仅供学习和研究使用，请勿用于非法用途。严禁复制、修改后进行售卖等商业行为。
2. 工程内所有直播源均来自网络，如有侵权，请联系作者删除。

## 📦 安装方式

### 使用 Go 安装（推荐）

如果你已安装 `Golang`，可通过以下命令快速安装：

```shell
go install github.com/rambollwong/rainbow-iptv-source-filter/cmd/rainbow-iptv-source-filterd@latest
```

### 使用预编译二进制文件

若未安装 Go，可前往 [GitHub Releases](https://github.com/rambollwong/rainbow-iptv-source-filter/releases) 下载适用于你系统的二进制文件。

## 🚀 使用方法

安装完成后，可通过以下命令启动程序：

```shell
# Linux 或 macOS
rainbow-iptv-source-filterd -c ./conf

# Windows
rainbow-iptv-source-filterd.exe -c ./conf
```

### 配置文件准备

运行前，请在 `./conf` 目录下创建名为 `config.yaml` 的配置文件，配置内容请参考下方 [配置文件说明](#配置文件说明)。

### 运行日志示例

若配置正确，程序将开始运行并输出日志，示例如下：

![](https://ramboll.wang/img/RainbowIPTVSourceFilter_log_snip.png)

在运行过程中，可能会出现红色的 `ERR` 日志，只要程序未退出，均为正常现象。  
程序运行结束后，将输出 `All Done.` 字样。

![](https://ramboll.wang/img/RainbowIPTVSourceFilter_log_alldone_snip.png)

## ⚙️ 配置文件说明

```yaml
programListSourceUrls: # 网络直播源列表，支持多个，同时支持`,m3u`和`.txt`格式
  - https://raw.githubusercontent.com/kimwang1978/collect-tv-txt/refs/heads/main/bbxx_lite.m3u
  - https://raw.githubusercontent.com/Guovin/iptv-api/gd/output/result.m3u
  - https://raw.githubusercontent.com/Guovin/iptv-api/refs/heads/gd/output/result.txt
  - https://raw.githubusercontent.com/yuanzl77/IPTV/main/live.m3u
programListSourceFileLocalPath: path/to/local/files # 本地直播源文件所在目录
outputFile: ./output/result.m3u # 输出文件路径，工具会根据文件后缀来确定输出文件格式，支持`.m3u`和`.txt`格式，默认为 `.m3u`
testPingMinLatency: 5000 # 每个节目单地址的最低访问延迟（单位：ms）
testLoadMinSpeed: 800 # 每个直播源的最低读取速度（单位：kb/s），低于该值的源将被过滤
retryTimes: 3 # 访问失败后的重试次数
customUA: # 自定义 User-Agent（可选）
parallelExecutorNum: 50 # 并发测试线程数，可根据电脑性能和网络带宽调整
groupList: # 自定义频道分组，仅测试定义在此处的频道
  - group: 央视 # 分组名称
    tvgName: # 频道列表（注意不要重复）
      - CCTV1,CCTV1综合 # 支持多频道名合并，通过这种方式兼容不同直播源的频道命名，最终以最左侧频道名输出到最终文件中
      - CCTV2,CCTV2财经
      - CCTV3,CCTV3综艺
      - CCTV4,CCTV4中文国际
      - CCTV5,CCTV5体育
      - CCTV5+,CCTV5+体育赛事
      - CCTV6,CCTV6电影
      - CCTV7,CCTV7国防军事,CCTV7军事
      - CCTV8,CCTV8电视剧
      - CCTV9,CCTV9纪录
      - CCTV10,CCTV10科教
      - CCTV11,CCTV11戏曲
      - CCTV12,CCTV12社会与法
      - CCTV13,CCTV13新闻
      - CCTV14,CCTV14少儿
      - CCTV15,CCTV15音乐
      - CCTV16,CCTV16-MST
      - CCTV17,CCTV17农业农村
      - CCTV4K,CCTV4KMXW
      - CCTV8K,CCTV8KMCP
  - group: 卫视
    tvgName:
      - 北京卫视
      - 江苏卫视
      - 浙江卫视
      - 湖南卫视
      - 东方卫视
      - 辽宁卫视
      - 天津卫视
      - 黑龙江卫视
      - 广东卫视
      - 深圳卫视
      - 山东卫视
      - 四川卫视
      - 安徽卫视
      - 东南卫视
      - 福建卫视
      - 贵州卫视
      - 云南卫视
      - 河南卫视
      - 重庆卫视
      - 湖北卫视
      - 江西卫视
      - 广西卫视
      - 河北卫视
      - 山西卫视
      - 陕西卫视
      - 海南卫视
      - 吉林卫视
      - 内蒙古卫视
      - 新疆卫视
      - 西藏卫视
      - 宁夏卫视
      - 甘肃卫视
      - 青海卫视
      - 厦门卫视
hostCustomUA: # 针对特定域名/地址的UA设置
  - mursor.ottiptv.cc -> okHttp/Mod-1.0.1

# 日志配置（如无特殊需求，建议保持默认）
# 详情请参考 rainbowlog：https://github.com/rambollwong/rainbowlog
rainbowlog:
  enable: true                    # enable logger
  level: INFO                    # default logger level
  label: ""                       # default logger label, if empty, will not record the label
  stack: false                    # whether print stack
  enableConsolePrinting: true     # whether print log record to console
  enableRainbowConsole: true      # whether using rainbow colors when printing to console
  timeFormat:                     # the time format of the time in each record, e.g. 'UNIX' or 'UNIXMS' or 'UNIXMICRO' or 'UNIXNANO' or '2006-01-02 15:04:05.000'
  sizeRollingFileConfig:
    enable: false                 # enable size rolling file
    logFilePath: ./log            # the path of log files
    logFileBaseName: rainbow.iptv.source.filter.log  # the base name of log file
    maxBackups: 10                # max log file backups, if it is negative, the file rotating will be disabled
    fileSizeLimit: 100M           # the max size of each log file, it is valid when MaxBackups is not negative
    encoder: txt                  # specify the log information format of the log file, 'txt' and 'json' supported.
  timeRollingFileConfig:
    enable: false                 # enable time rolling file
    logFilePath: ./log            # the path of log files
    logFileBaseName: rainbow.iptv.source.filter.log  # the base name of log file
    maxBackups: 7                 # max log file backups, if it is negative, the file rotating will be disabled
    rollingPeriod: DAY            # the rolling time period for rotating log file, e.g. 'YEAR' or 'MONTH' or 'DAY' or 'HOUR' or 'MINUTE' or 'SECOND'
    encoder: txt                  # specify the log information format of the log file, 'txt' and 'json' supported.
```

## 部分实现细节

1. 在测试过程中，本工具会自动过滤掉 URL 中包含 `audio` 关键词的源。这是因为此类源通常为音频流，而非视频直播流，不适用于本工具的目标场景。
2. ~~当前版本中，若某个源的 `tvg-name` 与 `title` 不一致，该源也会被过滤。此行为将在后续版本中调整，未来将统一以 `tvg-name` 作为匹配标准。~~
3. 所有频道名`tvg-name`都将被转换为大写，并去除`-`字符。

## 📬 联系我们

- 邮箱：`ramboll.wong@hotmail.com`
- Telegram 技术交流群：[点击加入](https://t.me/+EZ0us2YdjeE3YTk1)
- 博客：[Ramboll's Blog](https://ramboll.wang)

## 🙏 鸣谢

感谢以下仓库提供直播源数据支持：

- [kimwang1978/collect-txt](https://github.com/kimwang1978/collect-txt)
- [Guovin/iptv-api](https://github.com/Guovin/iptv-api)
- [yuanzl77/IPTV](https://github.com/yuanzl77/IPTV)
- [mursor1985/LIVE](https://github.com/mursor1985/LIVE)

## 💰 打赏支持

如果你喜欢这个项目，欢迎请作者喝杯柠檬水 ☕️，你的支持是我持续更新的动力！

- 微信支付：

<img src="https://ramboll.wang/img/wechat_pay.jpg" alt="WeChat Pay" width="200"/>

- 支付宝：

<img src="https://ramboll.wang/img/ali_pay.jpg" alt="Alipay" width="200"/>

