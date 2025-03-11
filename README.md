<h3 align="center">3x-ui Traffic Exporter</h3>
<div align="center">

[![GitHub stars](https://img.shields.io/github/stars/itning/3x-ui-traffic-exporter.svg?style=social&label=Stars)](https://github.com/itning/3x-ui-traffic-exporter/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/itning/3x-ui-traffic-exporter.svg?style=social&label=Fork)](https://github.com/itning/3x-ui-traffic-exporter/network/members)
[![GitHub watchers](https://img.shields.io/github/watchers/itning/3x-ui-traffic-exporter.svg?style=social&label=Watch)](https://github.com/itning/3x-ui-traffic-exporter/watchers)
[![GitHub followers](https://img.shields.io/github/followers/itning.svg?style=social&label=Follow)](https://github.com/itning?tab=followers)


</div>

<div align="center">

[![GitHub issues](https://img.shields.io/github/issues/itning/3x-ui-traffic-exporter.svg)](https://github.com/itning/3x-ui-traffic-exporter/issues)
[![GitHub license](https://img.shields.io/github/license/itning/3x-ui-traffic-exporter.svg)](https://github.com/itning/3x-ui-traffic-exporter/blob/master/LICENSE)
[![GitHub last commit](https://img.shields.io/github/last-commit/itning/3x-ui-traffic-exporter.svg)](https://github.com/itning/3x-ui-traffic-exporter/commits)
[![GitHub repo size in bytes](https://img.shields.io/github/repo-size/itning/3x-ui-traffic-exporter.svg)](https://github.com/itning/3x-ui-traffic-exporter)
[![Hits](https://hitcount.itning.com?u=itning&r=3x-ui-traffic-exporter)](https://github.com/itning/hit-count)

</div>

---

# 介绍

功能：将3x-ui中的流量信息上报给Prometheus

实现效果：

![](./pic/a.png)

在3x-ui中：

![](./pic/b.png)

# 使用

```shell
./3x-ui-traffic-exporter-linux-amd64 --web.listen-address=":9100" 
```

```text
# HELP email_download_bytes_total Total bytes downloaded by each email.
# TYPE email_download_bytes_total counter
email_download_bytes_total{email="60fvu4mn",enable="1"} 1.247432473107e+12
email_download_bytes_total{email="6g0ff2mh",enable="1"} 1.4792804919e+10
email_download_bytes_total{email="jk8120as",enable="1"} 2.725568336e+09
email_download_bytes_total{email="rda66ai2",enable="1"} 1.78623051e+08
# HELP email_upload_bytes_total Total bytes uploaded by each email.
# TYPE email_upload_bytes_total counter
email_upload_bytes_total{email="60fvu4mn",enable="1"} 2.24460727439e+11
email_upload_bytes_total{email="6g0ff2mh",enable="1"} 1.43013856e+08
email_upload_bytes_total{email="jk8120as",enable="1"} 1.33524713e+08
email_upload_bytes_total{email="rda66ai2",enable="1"} 4.6806666e+07
```

默认使用3x-ui sqlite数据库位置为：`/etc/x-ui/x-ui.db`

如果不在默认位置可以通过参数修改，例如`--db-path=/home/xui.db`

支持TLS：`--web.config.file=web-config.yml`

具体配置：[exporter-toolkit web-configuration](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md)

# 感谢

![JetBrains Logo (Main) logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)