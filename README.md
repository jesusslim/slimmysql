# slimmysql

a nice tool for mysql query

go语言下的mysql工具类 摈弃了ORM

contact me:755808379@qq.com

github地址:https://github.com/jesusslim/slimmysql

	使用：
	//初始化
	err = slimmysql.InitSqlDefault("user", "pass", "ip", "port", "db", "prefix", false) //last param true:check sql safe
	
	//More connections? 多个数据库连接
	err = slimysql.InitSql(1,"user", "pass", "ip", "port", "db", "prefix", false) //last param true:check sql safe
	//How to use? 切换数据库连接
	slimsql.SetConn(1)	
	
	if err != nil {
		this.Ctx.WriteString(err.Error())
	}
	
	slimsql := new(slimmysql.Sql)
	
	//查询
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



	例子
	//example:
	
	//function in a controller based on beego
	初始化
	err = slimmysql.InitSqlDefault("user", "pass", "ip", "port", "db", "prefix", false) //last param true:check sql safe
	
	if err != nil {
		this.Ctx.WriteString(err.Error())
	}
	
	slimsql := new(slimmysql.Sql)
	
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
	
	count, err := slimsql.Count("id")
	
	//mysql:Count: SELECT count(id) FROM `pre_students`  WHERE  ( id > 10385  AND ( ( status = '1'  AND acoin > 200 ) OR type = '2' ) AND  ( create_time >= 1431619200 AND create_time <= 1433088000 ) ) OR id >= 10575  OR nickname LIKE '%Jesus%'
	
	slimsql2 := new(slimmysql.Sql)
	
	tchs, err := slimsql2.Table("teachers").Where("status = 4").Page(1, 10).Fields("id,nickname").Select()
	
	//mysql:Select: SELECT id,nickname FROM `pre_teachers`  WHERE status = 4  LIMIT 0,10
	
	//Transaction
	sql := new(slimmysql.Sql)
	sql2 := new(slimmysql.Sql)
	sql.StartTrans()
	sql2.StartTrans()
	data := map[string]interface{}{
		"nickname": "tx",
	}
	data2 := map[string]interface{}{
		"nickname": "tx2",
	}
	sql.Table("students").Where("id=10605").Save(data)
	sql2.Table("students").Where("id=10606").Save(data2)
	sql.Rollback()
	sql2.Commit()
	
	
