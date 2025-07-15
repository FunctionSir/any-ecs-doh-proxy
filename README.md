<!--
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2025-04-15 19:57:34
 * @LastEditTime: 2025-07-15 19:04:02
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /any-ecs-doh-proxy/README.md
-->

# any-ecs-doh-proxy

DoH proxy, which can set ECS by city.

## How to deploy

Command is "FileNameOfTheBinVer /path/to/config/file".

Makesure you have a IPV4 SQLite DB with such schema:

``` sql
CREATE TABLE IPV4 (
    BEGIN TEXT, END TEXT,
    CODE TEXT, COUNTRY TEXT,
    PROVINCE TEXT, CITY TEXT
);
```

Code is "country" code.

It's easy to build from IP2Location LITE database.

Just do this:

``` sh
sqlite3 /path/to/db
```

```sql
CREATE TABLE IPV4 (
    BEGIN TEXT, END TEXT,
    CODE TEXT, COUNTRY TEXT,
    PROVINCE TEXT, CITY TEXT
);
.mode csv
.import /path/of/csv/format/IP2Location-LITE-IP-COUNTRY-REGION-CITY-Database
.exit
```

Infact, we provide one that we bulit, you can download it at [https://pubres.vioxt.eu.org/DoH/ipv4.db](https://pubres.vioxt.eu.org/DoH/ipv4.db).

Make it a service, or just use nohup, or even in tmux.

## How to config

Config:

``` ini
[options]
IpDb = ipv4.db
Listen = 127.0.0.1:8080
# Specify the homepage, if you don't set it, default will be used.
HomePage = example.html
# You don't need to specify Cert or Key if you just want run an HTTP version.
Cert = /path/to/cert/file
Key = /path/to/key/file
# You don't need to specify Proxy if you don't want to use it.
Proxy = socks5://127.0.0.1:9150
# This is only a example, but the upstream server must supports ECS.
Upstream = https://9.9.9.11/dns-query
# This specified a list of sites which need Dynamic Upstream Server feature.
DusNeeded = dus-needed.txt
```

"DusNeeded":

``` txt
.example.org.
```

Note: "." in prefix and suffix should be added! Aka: Don't change to "example.org" in this example!

## Report security issues

If you want to report a security issue, DO NOT sent it to GitHub, please send email to me, thanks.

My email: <functionsir@outlook.com>.

You can find my GPG public key at "keys.openpgp.org"

Fingerprint: 7B235AFE17F9EFECF613095D1DDAE7FE9D2EA01C.

## Suggestions and not-security-related bug reports

Just use Issues on GitHub. Thanks!

## Thanks

### Technical

This article: <https://taoshu.in/dns/diy-doh.html> is really good. Thanks to the author.

### IP Location Data Source

The provided SQLite database uses the IP2Location LITE database for [IP geolocation](https://lite.ip2location.com).
