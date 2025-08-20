# Rainbow IPTV Source Filter

![](https://ramboll.wang/img/IPTVSourceFilter_LOGO_SD.png)

[[中文]](./README_CN.md)

`Rainbow IPTV Source Filter` is a tool for filtering IPTV sources. It can detect and filter out unreachable or poor-quality sources, and merge the available sources into a specified output file.

## ⚠️ Notes

1. This tool is designed to perform local availability tests on live sources within the same network environment as the IPTV playback device, and to merge the filtered valid sources into a target file. Please do not use this tool in scenarios involving cloud servers or environments that do not match the playback device's network.
2. By default, the tool uses `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36` as the User-Agent (UA) for testing access. To modify it, please set a custom UA in the configuration file. If the default UA is used, you may need to adjust the player's UA settings to ensure compatibility when playing the filtered sources (e.g., Tianguang Yinying player).

## ⚠️ Disclaimer

1. This tool is for learning and research purposes only. Please do not use it for illegal purposes. Copying, modifying, and selling the tool commercially is strictly prohibited.
2. All live sources in the project come from the internet. If there is any infringement, please contact the author for removal.

## 📦 Installation

### Install with Go (Recommended)

If you have `Golang` installed, you can quickly install the tool with the following command:

```shell
go install github.com/rambollwong/rainbow-iptv-source-filter/cmd/rainbow-iptv-source-filterd@latest
```

### Use Precompiled Binary

If Go is not installed, you can download the binary file suitable for your system from [GitHub Releases](https://github.com/rambollwong/rainbow-iptv-source-filter/releases).

## 🚀 Usage

After installation, you can start the program with the following commands:

```shell
# Linux or macOS
rainbow-iptv-source-filterd -c ./conf

# Windows
rainbow-iptv-source-filterd.exe -c ./conf
```

### Configuration File Preparation

Before running, please create a configuration file named `config.yaml` in the `./conf` directory. For configuration details, please refer to the [Configuration File Description](#configuration-file-description) below.

### Running Log Example

If the configuration is correct, the program will start running and output logs, as shown below:

![](https://ramboll.wang/img/RainbowIPTVSourceFilter_log_snip.png)

During the running process, red `ERR` logs may appear. As long as the program does not exit, this is normal.  
After the program finishes running, it will output the message `All Done.`.

![](https://ramboll.wang/img/RainbowIPTVSourceFilter_log_alldone_snip.png)

## ⚙️ Configuration File Description

```yaml
programListSourceUrls: # List of network live sources, multiple sources supported, both `.m3u` and `.txt` formats are supported
  - https://raw.githubusercontent.com/kimwang1978/collect-tv-txt/refs/heads/main/bbxx_lite.m3u
  - https://raw.githubusercontent.com/Guovin/iptv-api/gd/output/result.m3u
  - https://raw.githubusercontent.com/Guovin/iptv-api/refs/heads/gd/output/result.txt
  - https://raw.githubusercontent.com/yuanzl77/IPTV/main/live.m3u
programListSourceFileLocalPath: path/to/local/files # Directory of local live source files
outputFile: ./output/result.m3u # Output file path. The tool will determine the output file format based on the file extension. Both `.m3u` and `.txt` formats are supported, with `.m3u` as the default.
testPingMinLatency: 5000 # Minimum access latency for each program list address (unit: ms)
testLoadMinSpeed: 800 # Minimum read speed for each live source (unit: kb/s), sources below this value will be filtered out
retryTimes: 3 # Number of retries after access failure
customUA: # Custom User-Agent (optional)
parallelExecutorNum: 50 # Number of concurrent test threads, adjustable based on computer performance and network bandwidth
groupList: # Custom channel groups, only channels defined here will be tested
  - group: 央视 # Group name
    tvgName: # Channel list (avoid duplicates)
      - CCTV1
      - CCTV2
      - CCTV3
      - CCTV4
      - CCTV5
      - CCTV5+
      - CCTV6
      - CCTV7
      - CCTV8
      - CCTV9
      - CCTV10
      - CCTV11
      - CCTV12
      - CCTV13
      - CCTV14
      - CCTV15
      - CCTV16
      - CCTV17
      - CCTV4K
      - CCTV8K
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

# Log Configuration (recommended to keep default if no special requirements)
# For details, please refer to rainbowlog: https://github.com/rambollwong/rainbowlog
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

## Implementation Details

1. During the testing process, the tool automatically filters out sources whose URLs contain the keyword `audio`. This is because such sources are typically audio streams rather than video live streams, which do not align with the intended use case of this tool.
2. ~~In the current version, if a source's `tvg-name` does not match its `title`, that source will also be filtered out. This behavior will be adjusted in future versions, where `tvg-name` will be used uniformly as the matching standard.~~
3. All channel names `tvg-name` will be converted to uppercase, and the `-` character will be removed.

## 📬 Contact Us

- Email: `ramboll.wong@hotmail.com`
- Telegram Technical Discussion Group: [Join Now](https://t.me/+EZ0us2YdjeE3YTk1)
- Blog：[Ramboll's Blog](https://ramboll.wang)

## 🙏 Acknowledgments

Thanks to the following repositories for providing live source data support:

- [kimwang1978/collect-txt](https://github.com/kimwang1978/collect-txt)
- [Guovin/iptv-api](https://github.com/Guovin/iptv-api)
- [yuanzl77/IPTV](https://github.com/yuanzl77/IPTV)
- [mursor1985/LIVE](https://github.com/mursor1985/LIVE)

## 💰 Support with a Donation

If you like this project, feel free to buy the author a cup of lemonade ☕️. Your support is my motivation for continuous updates!

- WeChat Pay: 

<img src="https://ramboll.wang/img/wechat_pay.jpg" alt="WeChat Pay" width="200"/>

- Alipay: 

<img src="https://ramboll.wang/img/ali_pay.jpg" alt="Alipay" width="200"/>

