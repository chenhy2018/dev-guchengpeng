qiniumail manual
==================

## 命令行

```
qiniumail [-f <mail.conf>] -to <MailTo> -subject <MailSubject> <mail.body> [<maildata.mjson>]
cat <maildata.mjson> | qiniumail [-f <mail.conf>] -to <MailTo> -subject <MailSubject> <mail.body> -
```

其中：

* `<mail.conf>`: 发送邮件所需的配置文件。默认为 $HOME/.qiniumail/mail.conf。配置文件格式详后。
* `<mail.body>`: 邮件正文内容的模版文件。
* `<maildata.mjson>`: 邮件数据文件，可以被 `<mail.body>`、`-to <MailTo>`、`-subject <MailSubject>` 等所引用。`<maildata.mjson>` 是一个多行文本，每行是一个合法的 json。
* `<MailTo>`: 邮件的收件人。可以引用 `<maildata.mjson>` 的数据。
* `<MailSubject>`: 邮件标题。可以引用 `<maildata.mjson>` 的数据。

