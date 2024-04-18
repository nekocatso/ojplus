# Details

Date : 2024-04-17 20:52:04

Directory /projects/alarm

Total : 47 files,  7583 codes, 3 comments, 685 blanks, all 8271 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [cmd/alarm/main.go](/cmd/alarm/main.go) | Go | 184 | 0 | 22 | 206 |
| [config.toml](/config.toml) | TOML | 17 | 3 | 3 | 23 |
| [go.mod](/go.mod) | Go Module File | 62 | 0 | 5 | 67 |
| [go.sum](/go.sum) | Go Checksum File | 209 | 0 | 1 | 210 |
| [internal/config/config.go](/internal/config/config.go) | Go | 54 | 0 | 10 | 64 |
| [internal/pkg/listener/listener.go](/internal/pkg/listener/listener.go) | Go | 159 | 0 | 20 | 179 |
| [internal/pkg/listener/listener_test.go](/internal/pkg/listener/listener_test.go) | Go | 89 | 0 | 4 | 93 |
| [internal/pkg/listener/messageparse.go](/internal/pkg/listener/messageparse.go) | Go | 86 | 0 | 6 | 92 |
| [internal/pkg/listenerpool/listenerpool.go](/internal/pkg/listenerpool/listenerpool.go) | Go | 600 | 0 | 52 | 652 |
| [internal/pkg/mail/mail.go](/internal/pkg/mail/mail.go) | Go | 118 | 0 | 20 | 138 |
| [internal/pkg/messagequeue/messagequeue.go](/internal/pkg/messagequeue/messagequeue.go) | Go | 183 | 0 | 28 | 211 |
| [internal/pkg/messagequeue/messagequeue_test.go](/internal/pkg/messagequeue/messagequeue_test.go) | Go | 264 | 0 | 25 | 289 |
| [internal/pkg/rule/ping.go](/internal/pkg/rule/ping.go) | Go | 366 | 0 | 60 | 426 |
| [internal/pkg/rule/rule.go](/internal/pkg/rule/rule.go) | Go | 29 | 0 | 5 | 34 |
| [internal/pkg/rule/tcp.go](/internal/pkg/rule/tcp.go) | Go | 456 | 0 | 60 | 516 |
| [internal/utils/handler.go](/internal/utils/handler.go) | Go | 20 | 0 | 4 | 24 |
| [internal/utils/validator.go](/internal/utils/validator.go) | Go | 16 | 0 | 4 | 20 |
| [internal/web/controllers/account.go](/internal/web/controllers/account.go) | Go | 345 | 0 | 13 | 358 |
| [internal/web/controllers/alarm.go](/internal/web/controllers/alarm.go) | Go | 317 | 0 | 11 | 328 |
| [internal/web/controllers/asset.go](/internal/web/controllers/asset.go) | Go | 391 | 0 | 14 | 405 |
| [internal/web/controllers/auth.go](/internal/web/controllers/auth.go) | Go | 216 | 0 | 11 | 227 |
| [internal/web/controllers/log.go](/internal/web/controllers/log.go) | Go | 303 | 0 | 12 | 315 |
| [internal/web/controllers/response.go](/internal/web/controllers/response.go) | Go | 85 | 0 | 6 | 91 |
| [internal/web/controllers/rule.go](/internal/web/controllers/rule.go) | Go | 407 | 0 | 14 | 421 |
| [internal/web/controllers/util.go](/internal/web/controllers/util.go) | Go | 9 | 0 | 3 | 12 |
| [internal/web/forms/account.go](/internal/web/forms/account.go) | Go | 144 | 0 | 14 | 158 |
| [internal/web/forms/alarm.go](/internal/web/forms/alarm.go) | Go | 112 | 0 | 15 | 127 |
| [internal/web/forms/asset.go](/internal/web/forms/asset.go) | Go | 202 | 0 | 17 | 219 |
| [internal/web/forms/auth.go](/internal/web/forms/auth.go) | Go | 41 | 0 | 9 | 50 |
| [internal/web/forms/form.go](/internal/web/forms/form.go) | Go | 52 | 0 | 7 | 59 |
| [internal/web/forms/log.go](/internal/web/forms/log.go) | Go | 147 | 0 | 15 | 162 |
| [internal/web/forms/rule.go](/internal/web/forms/rule.go) | Go | 185 | 0 | 18 | 203 |
| [internal/web/logs/user.go](/internal/web/logs/user.go) | Go | 30 | 0 | 7 | 37 |
| [internal/web/models/account.go](/internal/web/models/account.go) | Go | 17 | 0 | 3 | 20 |
| [internal/web/models/alarm.go](/internal/web/models/alarm.go) | Go | 18 | 0 | 4 | 22 |
| [internal/web/models/asset.go](/internal/web/models/asset.go) | Go | 28 | 0 | 5 | 33 |
| [internal/web/models/cache.go](/internal/web/models/cache.go) | Go | 24 | 0 | 5 | 29 |
| [internal/web/models/db.go](/internal/web/models/db.go) | Go | 33 | 0 | 5 | 38 |
| [internal/web/models/log.go](/internal/web/models/log.go) | Go | 26 | 0 | 4 | 30 |
| [internal/web/models/rule.go](/internal/web/models/rule.go) | Go | 41 | 0 | 6 | 47 |
| [internal/web/services/account.go](/internal/web/services/account.go) | Go | 166 | 0 | 16 | 182 |
| [internal/web/services/alarm.go](/internal/web/services/alarm.go) | Go | 208 | 0 | 16 | 224 |
| [internal/web/services/asset.go](/internal/web/services/asset.go) | Go | 437 | 0 | 33 | 470 |
| [internal/web/services/auth.go](/internal/web/services/auth.go) | Go | 87 | 0 | 18 | 105 |
| [internal/web/services/log.go](/internal/web/services/log.go) | Go | 171 | 0 | 16 | 187 |
| [internal/web/services/rule.go](/internal/web/services/rule.go) | Go | 380 | 0 | 32 | 412 |
| [internal/web/services/util.go](/internal/web/services/util.go) | Go | 49 | 0 | 7 | 56 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)