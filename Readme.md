# route53_exporter
Exposes prometheus endpoint with route53 specific metrics from given cluster.Currently generated metrics:

```
aws_route53_hostedzone_resourcerecordsetcount{account=aws_account",hostedzoneid="/hostedzone/<UNIQUEID>",name="a3.dev.yourdomain.work.",privateZone="false"} 2
aws_route53_hostedzone_resourcerecordsetlimit{account="aws_account",hostedzoneid="/hostedzone/<UNIQUEID>",name="a3.dev.yourdomain.work.",privateZone="true"} 10000
```

| Tags available | Values example|
|--|--|
|account|"aws_account_name"|
|name|"hostname.xyz.com"|
|hostedzoneid|"route53 hostedzoneId"|
|privateZone|boolean true false|
