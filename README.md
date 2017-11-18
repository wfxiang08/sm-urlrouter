# httprouter
* 基本支持yii中的url的各种pattern的处理
	* http://www.yiiframework.com/doc/guide/1.1/en/topics.url
* 用途:
 	* 可以通过rpc服务将yii中的UrlManager中的ParseRequest接口进行重写，提升url resolve的效率
 	* 可以在日志分析时将各种带有参数的url解析成为对应的action, 然后便于统计分析