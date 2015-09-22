# slimmysql

a nice tool for mysql query

go语言下的mysql工具类 摈弃了ORM

contact me:755808379@qq.com

github地址:https://github.com/jesusslim/slimmysql

使用：

完整例子
full example:

[Slimgotest][slimgotestlink]


Init(reg connections)
	
	conf, _ := goconfig.LoadConfigFile("./conf/db.ini")
	conf_sq := "local"
	slimmysql.RegisterConnectionDefault(conf.MustBool(conf_sq, "rwseparate"), conf.MustValue(conf_sq, "host"), conf.MustValue(conf_sq, "port"), conf.MustValue(conf_sq, "db"), conf.MustValue(conf_sq, "user"), conf.MustValue(conf_sq, "pass"), conf.MustValue(conf_sq, "prefix"), false)
	
the conf shoud be like this:

配置文件参考

	[local]
	rwseparate = true
	user = root
	pass = root
	host = 127.0.0.1,127.0.0.1,127.0.0.1
	port = 3307
	db = testgo,testgo2,testgo3
	prefix = go_

Multi hosts means read/write separate,use the first host as master for write,others for read.

多个host表示主从读写分离 第一个地址为主库 其他为从库

Then get a Instance when you want to use it.

使用时使用该方法获取一个实例

	slimsql, err := slimmysql.NewSqlInstanceDefault()
	

查询 condition

采用map[string]interface{}作为查询条件传入
map的key代表字段 value代表指
like/>/</between等等都在字段名称后面加上__like等
例如 nickname like : key=nickname__like
	
	默认为＝
	like:__like
	>:__gt
	>=:__egt
	<:__lt
	<=:__elt
	!=:__neq
	between:__between
	in:__in
	notin:__notin
	isnull:__isnull
	isnotnull:__isnotnull
	
key=relation表示and、or等 默认为and
如果需嵌套condition 则key="__" value=map[string]interfcae{}
	
例子：
		
		condition := map[string]interface{}{
		"relation":       "or",
		"nickname__like": "Jesus",
		"_": map[string]interface{}{
			"id__gt": 10385,
			"_": map[string]interface{}{
				"relation": "or",
				"_": map[string]interface{}{
					"status":    1,
					"acoin__gt": 200,
				},
				"type": 2,
			},
			"create_time__between": "1431619200|1433088000",
		},
		"id__egt": 10575,
	}
	
	stds, err := slimsql.Table("students").Where(condition).Page(1, 10).Order("id desc").Fields("id,nickname,nickname_cn,acoin,mobile").Select()
	
	//mysql:Select: SELECT id,nickname,nickname_cn,acoin,mobile FROM `pre_students`  WHERE  ( id > 10385  AND ( ( status = '1'  AND acoin > 200 ) OR type = '2' ) AND  ( create_time >= 1431619200 AND create_time <= 1433088000 ) ) OR id >= 10575  OR nickname LIKE '%Jesus%'  ORDER BY id desc  LIMIT 0,10
	
函数：
	
	where
	group
	having
	order
	page
	pk 设置主键
	join ：slimsql.Table("A").join("left join B on A.id = B.aid")
	count
	find 查询单个 返回一个map
	select 返回一个map数组
	getField 与select类似 返回一个value为map的map 以getfield中第一个字段作为key
	setinc 增长
	save update
	add insert
	delete
	
事务支持
	
	starttrans
	commit
	rollback
	
锁表等
	
	Lock
	Unlock
	LockRow(forupdate)
	
	ping 测试连接状态
	
	clear 清除slimmysql.Sql对象中的值（不清楚的情况下可重复使用 例如select之后 直接调用count可获得数量 而不需要重新传入condition等）
	
具体可参考gowalker：https://gowalker.org/github.com/jesusslim/slimmysql
	
	
	
	
[slimgotestlink]:https://github.com/jesusslim/slimgotest