package slimmysql

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
)

const VERSION = "2.1.0"

type Conns struct {
	sqlDB      *sql.DB   //master db for write | 主库
	sqlDBSub   []*sql.DB //dbs for read  | 从
	rwSeparate bool      //use read,write separate | 使用读写分离
	safeMode   bool      //safe for sql format | sql注入安全
	rwSepTag   int       //tag of dbs for read order | 读库顺序标记
	subCount   int       //count of dbs for read | 读库数量
	prefix     string    //prefix of tables | 数据库前缀
}

var conns map[string]*Conns //all connections

func init() {
	conns = make(map[string]*Conns)
}

/**
 * [RegisterConnectionDefault Reg the default connection]
 * @param {[bool]} rwSep                     bool          [use RWseparate]
 * @param {[string]} hosts                     [hosts for exp:192.168.0.1,192.168.0.2]
 * @param {[string]} port                      [port for exp:3306,3307]
 * @param {[string]} dbname                    [dbnames]
 * @param {[string]} user                      [users]
 * @param {[string]} pass                      string        [passwords]
 * @param {[string]} pre                       string        [prefix]
 * @param {[bool]} safe                      bool          [safemode]
 * @param {[int]} maxOpenConnsNmaxIdleConns ...int        [maxOpenConns and maxIdleConns for exp:2000,1000]
 */
func RegisterConnectionDefault(rwSep bool, hosts, port, dbname, user, pass string, pre string, safe bool, maxOpenConnsNmaxIdleConns ...int) {
	RegisterConnection("default", rwSep, hosts, port, dbname, user, pass, pre, safe, maxOpenConnsNmaxIdleConns...)
}

/**
 * [RegisterConnection Reg a connection]
 * @param {[string]} name                      string        [connection name]
 * @param {[bool]} rwSep                     bool          [use RWseparate]
 * @param {[string]} hosts                     [hosts for exp:192.168.0.1,192.168.0.2]
 * @param {[string]} port                      [port for exp:3306,3307]
 * @param {[string]} dbname                    [dbnames]
 * @param {[string]} user                      [users]
 * @param {[string]} pass                      string        [passwords]
 * @param {[string]} pre                       string        [prefix]
 * @param {[bool]} safe                      bool          [safemode]
 * @param {[int]} maxOpenConnsNmaxIdleConns ...int        [maxOpenConns and maxIdleConns for exp:2000,1000]
 */
func RegisterConnection(name string, rwSep bool, hosts, port, dbname, user, pass string, pre string, safe bool, maxOpenConnsNmaxIdleConns ...int) {
	conn := &Conns{
		rwSeparate: rwSep,
		safeMode:   safe,
		rwSepTag:   0,
		subCount:   0,
		prefix:     pre,
	}
	maxOpenConns := 2000
	maxIdleConns := 1000
	if len(maxOpenConnsNmaxIdleConns) == 2 {
		maxOpenConns = maxOpenConnsNmaxIdleConns[0]
		maxIdleConns = maxOpenConnsNmaxIdleConns[1]
	} else if len(maxOpenConnsNmaxIdleConns) == 1 {
		maxOpenConns = maxOpenConnsNmaxIdleConns[0]
	}
	if rwSep {
		hostsArr := strings.Split(hosts, ",")
		hostsNum := len(hostsArr)
		if hostsNum == 0 {
			panic("No hosts set.")
		}
		portArr := strings.Split(port, ",")
		if len(portArr) == 0 {
			panic("No ports set.")
		}
		dbnameArr := strings.Split(dbname, ",")
		if len(dbnameArr) == 0 {
			panic("No db set.")
		}
		userArr := strings.Split(user, ",")
		if len(userArr) == 0 {
			panic("No user set.")
		}
		passArr := strings.Split(pass, ",")
		if len(passArr) == 0 {
			panic("No password set.")
		}
		//reg master
		dbmaster, err := conn.initSql(hostsArr[0], portArr[0], dbnameArr[0], userArr[0], passArr[0], maxOpenConns, maxIdleConns)
		if err != nil {
			panic("Register Mysql Connection Failed.")
		}
		conn.sqlDB = dbmaster
		//reg sub
		if len(hostsArr) > 1 {
			//has sub
			for key, hostTemp := range hostsArr[1:] {
				k := key + 1
				var portTemp, dbnameTemp, userTemp, passTemp string
				var ok bool
				if ok = k < len(portArr); !ok {
					portTemp = portArr[len(portArr)-1]
				} else {
					portTemp = portArr[k]
				}
				if ok = k < len(dbnameArr); !ok {
					dbnameTemp = dbnameArr[len(dbnameArr)-1]
				} else {
					dbnameTemp = dbnameArr[k]
				}
				if ok = k < len(userArr); !ok {
					userTemp = userArr[len(userArr)-1]
				} else {
					userTemp = userArr[k]
				}
				if ok = k < len(passArr); !ok {
					passTemp = passArr[len(passArr)-1]
				} else {
					passTemp = passArr[k]
				}
				dbsub, err := conn.initSql(hostTemp, portTemp, dbnameTemp, userTemp, passTemp, maxOpenConns, maxIdleConns)
				if err != nil {
					panic("Register Mysql Connection Failed.")
				}
				conn.sqlDBSub = append(conn.sqlDBSub, dbsub)
				conn.subCount++
			}
		}
	} else {
		dbmaster, err := conn.initSql(hosts, port, dbname, user, pass, maxOpenConns, maxIdleConns)
		if err != nil {
			panic("Register Mysql Connection Failed.")
		}
		conn.sqlDB = dbmaster
	}
	conns[name] = conn
}

func (this *Conns) initSql(host, port, dbname, user, pass string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	slimSqlLog("Init", "user:"+user+" pass:"+pass+" host:"+host+" port:"+port+" dbname:"+dbname)
	var sqlDB *sql.DB
	var err error
	sqlDB, err = sql.Open("mysql", user+":"+pass+"@tcp("+host+":"+port+")/"+dbname+"?charset=utf8")
	if err != nil {
		return nil, err
	}
	if maxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(maxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(2000)
	}
	if maxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(maxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(1000)
	}
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	return sqlDB, nil
}

/**
 * 获取写库sql.DB
 */
func (this *Conns) getDbW() *sql.DB {
	return this.sqlDB
}

/**
 * 获取读库sql.DB
 */
func (this *Conns) getDbR() *sql.DB {
	if this.rwSeparate {
		//无从库 走主库
		if this.subCount == 0 {
			return this.sqlDB
		}
		//按顺序分配从库
		if this.rwSepTag < this.subCount {
			dbUsed := this.sqlDBSub[this.rwSepTag]
			this.rwSepTag++
			return dbUsed
		} else {
			this.rwSepTag = 1
			return this.sqlDBSub[0]
		}
	} else {
		return this.sqlDB
	}
}

/**
 * [NewSqlInstanceDefault Get a new instance for user]
 */
func NewSqlInstanceDefault() (*Sql, error) {
	return NewSqlInstance("default")
}

/**
 * [NewSqlInstance Get a new instance for user]
 * @param {[string]} name string [connection name]
 * @return (*Sql, error)
 */
func NewSqlInstance(name string) (*Sql, error) {
	conn, ok := conns[name]
	if !ok {
		return nil, errors.New("Connection " + name + " not found.")
	}
	sqlInstance := &Sql{
		conn_id:    name,
		connection: conn,
	}
	return sqlInstance.Clear(), nil
}

type Sql struct {
	conn_id      string //Which connection to use
	connection   *Conns //connection
	mustMaster   bool   //must use db master
	fieldsSql    string
	tableName    string
	conditionSql string //sql in where
	joinSql      string
	groupSql     string
	havingSql    string
	orderSql     string
	pageSql      string
	tx           *sql.Tx //Transaction
	pkSql        string  //primary key
	forupdate    bool    //for update
}

func (this *Sql) GetConnectionName() string {
	return this.conn_id
}

func (this *Sql) MustMaster(mustMaster bool) *Sql {
	this.mustMaster = mustMaster
	return this
}

/**
 * Get the sql.DB from db write
 */
func (this *Sql) getDbW() *sql.DB {
	return this.connection.getDbW()
}

/**
 * Get the sql.DB from db read
 */
func (this *Sql) getDbR() *sql.DB {
	if this.mustMaster {
		return this.connection.getDbW()
	} else {
		return this.connection.getDbR()
	}
}

/**
 * Set table name,use prefix
 */
func (this *Sql) Table(tablename string) *Sql {
	this.tableName = this.connection.prefix + this.safeInSql(tablename, 1)
	return this
}

/**
 * Set table name,no prefix
 */
func (this *Sql) TrueTable(tablename string) *Sql {
	this.tableName = this.safeInSql(tablename, 1)
	return this
}

/**
 * Set primary key
 */
func (this *Sql) Pk(pk string) *Sql {
	this.pkSql = pk
	return this
}

/**
 * Set fields we want select
 * example:"id,title"
 */
func (this *Sql) Fields(filed string) *Sql {
	this.fieldsSql = this.safeInSql(filed, 1)
	return this
}

/**
 * Join table
 * example:"INNER JOIN B on A.id = B.aid"
 */
func (this *Sql) Join(joinsql string) *Sql {
	this.joinSql = this.safeInSql(joinsql, 2)
	return this
}

/**
 * group by
 */
func (this *Sql) Group(g string) *Sql {
	this.groupSql = this.safeInSql(g, 1)
	return this
}

/**
 * having after group by
 * example:"age > 24"
 */
func (this *Sql) Having(h string) *Sql {
	this.havingSql = this.safeInSql(h, 2)
	return this
}

/**
 * order by
 * example:"id desc"
 */
func (this *Sql) Order(order string) *Sql {
	this.orderSql = this.safeInSql(order, 2)
	return this
}

/**
 * page limit
 */
func (this *Sql) Page(page int, pagesize int) *Sql {
	offset := (page - 1) * pagesize
	this.pageSql = " LIMIT " + strconv.Itoa(offset) + "," + strconv.Itoa(pagesize)
	return this
}

/**
 * where
 * @param condition:string/map
 */
func (this *Sql) Where(condition interface{}) *Sql {
	switch condition.(type) {
	case string:
		this.conditionSql = this.safeInSql(condition.(string), 2)
		break
	case map[string]interface{}:
		this.conditionSql = this.convertCondition(condition.(map[string]interface{}))
		break
	}
	return this
}

/**
 * convert condition to sql
 */
func (this *Sql) convertCondition(condition map[string]interface{}) string {
	sql := ""
	join := "AND"
	if condition["relation"] != nil && strings.ToUpper(condition["relation"].(string)) == "OR" {
		join = "OR"
	}
	i := 0
	join_sql := " "
	var temp_key string
	var sign string
	for k, v := range condition {
		if k == "relation" {
			continue
		}
		if i == 0 {
			join_sql = " "
		} else {
			join_sql = " " + join + " "
		}
		if k == "_" {
			sql += join_sql + "(" + this.convertCondition(v.(map[string]interface{})) + ")"
			i++
			continue
		}
		temp := this.convertValue2String(v)
		if len(temp) > 0 {
			if strings.Contains(k, "__") {
				keys := strings.Split(k, "__")
				temp_key = keys[0]
				sign = keys[1]
			} else {
				temp_key = k
				sign = "="
			}
			switch sign {
			case "=":
				sql += join_sql + temp_key + " = " + "'" + temp + "' "
				break
			case "neq":
				sql += join_sql + temp_key + " != " + "'" + temp + "' "
				break
			case "gt":
				sql += join_sql + temp_key + " > " + "'" + temp + "' "
				break
			case "egt":
				sql += join_sql + temp_key + " >= " + "'" + temp + "' "
				break
			case "lt":
				sql += join_sql + temp_key + " < " + "'" + temp + "' "
				break
			case "elt":
				sql += join_sql + temp_key + " <= " + "'" + temp + "' "
				break
			case "between":
				temp_v := strings.Split(temp, "|")
				sql += join_sql + " ( " + temp_key + " BETWEEN '" + temp_v[0] + "' AND '" + temp_v[1] + "' ) "
				break
			case "in":
				sql += join_sql + temp_key + " IN (" + temp + ") "
				break
			case "notin":
				sql += join_sql + temp_key + " NOT IN (" + temp + ") "
				break
			case "like":
				sql += join_sql + temp_key + " LIKE '%" + temp + "%' "
				break
			case "isnull":
				sql += join_sql + " ISNULL(" + temp_key + ") "
				break
			case "isnotnull":
				sql += join_sql + " " + temp_key + " IS NOT NULL "
				break
			default:
				sql += join_sql + " 1 = 1 "
				break
			}
			i++
		}
	}
	return sql
}

/**
 * return the final sql str
 * @param isSelect true select false count
 */
func (this *Sql) GetSql(isSelect bool) string {
	sql := ""
	if isSelect {
		sql += "SELECT "
	} else {
		sql += "SELECT count("
	}
	if len(this.fieldsSql) > 0 {
		sql += this.fieldsSql
	} else {
		sql += "*"
	}
	if isSelect {

	} else {
		sql += ")"
	}
	sql += " FROM `" + this.tableName + "` "
	if len(this.joinSql) > 0 {
		sql += this.joinSql + " "
	}
	if len(this.conditionSql) > 0 {
		sql += " WHERE " + this.conditionSql
	}
	if len(this.groupSql) > 0 {
		sql += " GROUP BY " + this.groupSql
		if len(this.havingSql) > 0 {
			sql += " HAVING " + this.havingSql
		}
	}
	if len(this.orderSql) > 0 {
		sql += " ORDER BY " + this.orderSql
	}
	if len(this.pageSql) > 0 {
		sql += " " + this.pageSql
	}
	return sql
}

/**
 * select one return a single map
 */
func (this *Sql) Find(id interface{}) (map[string]string, error) {
	if len(this.pageSql) > 0 {
		this.pageSql = ""
	}
	if id != nil {
		sel_id := this.convertValue2String(id)
		if len(this.pkSql) > 0 {
			this.conditionSql = this.pkSql + " = '" + sel_id + "'"
		} else {
			this.conditionSql = "id = '" + sel_id + "'"
		}
	}
	sqlstr := this.GetSql(true) + " limit 1"
	var rows *sql.Rows
	var err error
	if this.tx != nil {
		if this.forupdate == true {
			sqlstr += " FOR UPDATE"
		}
		rows, err = this.tx.Query(sqlstr)
	} else {
		rows, err = this.getDbR().Query(sqlstr)
	}
	//slimSqlLog("Find", sqlstr)
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		columns, _ := rows.Columns()
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		record := make(map[string]string)
		for j := range values {
			scanArgs[j] = &values[j]
		}
		rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}
		return record, nil
	}
	return nil, nil
}

/**
 * base select func
 * @pk true:return a map,first field is key|false:reuturn an array
 */
func (this *Sql) baseSelect(pk bool) (map[string](map[string]string), []map[string]string, error) {
	sqlstr := this.GetSql(true)
	var rows *sql.Rows
	var err error
	if this.tx != nil {
		if this.forupdate == true {
			sqlstr += " FOR UPDATE"
		}
		rows, err = this.tx.Query(sqlstr)
	} else {
		rows, err = this.getDbR().Query(sqlstr)
	}
	//slimSqlLog("Select", sqlstr)
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return nil, nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		slimSqlLog("ERROR", err.Error())
		return nil, nil, err
	}
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}
	records := make(map[string](map[string]string))
	var records_arr []map[string]string
	for rows.Next() {
		record := make(map[string]string)
		rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			} else {
				record[columns[i]] = ""
			}
		}
		if pk {
			records[record[columns[0]]] = record
		} else {
			records_arr = append(records_arr, record)
		}
	}
	return records, records_arr, nil
}

/**
 * select func,return array
 */
func (this *Sql) Select() ([]map[string]string, error) {
	_, r, err := this.baseSelect(false)
	return r, err
}

/**
 * select func,return map,key is the first field
 */
func (this *Sql) GetField(fields string) (map[string](map[string]string), error) {
	this.fieldsSql = fields
	r, _, err := this.baseSelect(true)
	return r, err
}

/**
 * Count
 */
func (this *Sql) Count(filed string) (int, error) {
	if len(filed) > 0 {
		this.fieldsSql = this.convertValue2String(filed)
	} else {
		this.fieldsSql = "*"
	}
	this.orderSql = ""
	this.pageSql = ""
	sqlstr := this.GetSql(false)
	//slimSqlLog("Count", sqlstr)
	var rows *sql.Rows
	var err error
	if this.tx != nil {
		rows, err = this.tx.Query(sqlstr)
	} else {
		rows, err = this.getDbR().Query(sqlstr)
	}
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return 0, err
	}
	defer rows.Close()
	scanArgs := make([]interface{}, 1)
	values := make([]interface{}, 1)
	scanArgs[0] = &values[0]
	if rows.Next() {
		rows.Scan(scanArgs...)
		var count int
		count, err = strconv.Atoi(string(values[0].([]byte)))
		if err != nil {
			return 0, err
		}
		return count, nil
	}
	return 0, nil
}

/**
 * Filter for sql safe
 * @sql
 * @level 1:space 2:;and\n 3:1+2
 */
func (this *Sql) safeInSql(sql string, level int) string {
	if this.connection.safeMode {
		if level == 1 {
			return strings.Replace(sql, " ", "", -1)
		} else if level == 2 {
			return strings.Replace(strings.Replace(sql, ";", "", -1), "\n", "", -1)
		} else {
			return strings.Replace(strings.Replace(strings.Replace(sql, ";", "", -1), "\n", "", -1), " ", "", -1)
		}
	} else {
		return sql
	}
}

/**
 * Convert value to string with sql safe
 * @v value
 */
func (this *Sql) convertValue2String(v interface{}) string {
	var r string
	switch v.(type) {
	case string:
		r = this.safeInSql(v.(string), 1)
		break
	case int:
		r = strconv.Itoa(v.(int))
		break
	case int32:
		r = strconv.FormatInt(int64(v.(int32)), 10)
		break
	case int64:
		r = strconv.FormatInt(v.(int64), 10)
		break
	case float32:
		r = strconv.FormatFloat(float64(v.(float32)), 'f', -1, 10)
		break
	case float64:
		r = strconv.FormatFloat(v.(float64), 'f', -1, 10)
		break
	}
	return r
}

/**
 * Insert / convert ""/'' maybe a bug.
 * @param map of data
 * @return id,err
 */
func (this *Sql) Add(data map[string]interface{}) (int64, error) {
	var columns []string
	var values []string
	for k, v := range data {
		columns = append(columns, k)
		tmp_v := this.convertValue2String(v)
		values = append(values, "\""+tmp_v+"\"")
	}
	sqlstr := " INSERT INTO `" + this.tableName + "` " + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.Join(values, ",") + ") "
	//slimSqlLog("Insert", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = this.getDbW().Exec(sqlstr)
	}
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return 0, err
	} else {
		id, err := r.LastInsertId()
		if err != nil {
			return 0, err
		} else {
			return id, nil
		}
	}
}

/**
 * Update / convert ""/'' maybe a bug.
 * @param  map data
 * @return effect_colunms,err
 */
func (this *Sql) Save(data map[string]interface{}) (int64, error) {
	sqlstr := " UPDATE " + this.tableName + " "
	setsql := " SET "
	count := 0
	var pk string
	if len(this.pkSql) > 0 {
		pk = this.pkSql
	} else {
		pk = "id"
	}
	for k, v := range data {
		if k == pk {
			continue
		}
		if count > 0 {
			setsql += ", " + k + "=\"" + this.convertValue2String(v) + "\" "
		} else {
			setsql += " " + k + "=\"" + this.convertValue2String(v) + "\" "
		}
		count++
	}
	if len(this.conditionSql) > 0 {
		//do not need pk

	} else {
		this.conditionSql = pk + " = \"" + this.convertValue2String(data[pk]) + "\" "
	}
	sqlstr += setsql + " WHERE " + this.conditionSql
	//slimSqlLog("Update", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = this.getDbW().Exec(sqlstr)
	}
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return 0, err
	} else {
		num, err := r.RowsAffected()
		if err != nil {
			return 0, err
		} else {
			return num, nil
		}
	}
}

/**
 * SetInc
 * @param  fieldname string
 * @param  value int
 * @return rows,error
 */
func (this *Sql) SetInc(field string, value int) (int64, error) {
	sqlstr := " UPDATE " + this.tableName + " "
	sqlstr += " SET "
	field_filted := this.convertValue2String(field)
	sqlstr += field_filted + " = " + field_filted + " + " + this.convertValue2String(value)
	sqlstr += " WHERE "
	sqlstr += this.conditionSql
	//slimSqlLog("SETINC", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = this.getDbW().Exec(sqlstr)
	}
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return 0, err
	} else {
		num, err := r.RowsAffected()
		if err != nil {
			return 0, err
		} else {
			return num, nil
		}
	}
}

/**
 * Delete
 * @return effect,error
 */
func (this *Sql) Delete() (int64, error) {
	sqlstr := " DELETE FROM " + this.tableName + " WHERE " + this.conditionSql
	//slimSqlLog("Delete", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = this.getDbW().Exec(sqlstr)
	}
	if err != nil {
		slimSqlLog("ERROR", sqlstr)
		return 0, err
	} else {
		num, err := r.RowsAffected()
		if err != nil {
			return 0, err
		} else {
			return num, nil
		}
	}
}

func (this *Sql) StartTrans() (bool, string) {
	if this.tx != nil {
		//slimSqlLog("Tx", "Start new tx failed because tx is not null.")
		return false, "Tx not null!"
	} else {
		var err error
		this.tx, err = this.getDbW().Begin()
		if err != nil {
			slimSqlLog("Tx", "Start new tx failed because "+err.Error())
			return false, err.Error()
		} else {
			//slimSqlLog("Tx", "Start new tx success.")
			return true, ""
		}
	}
}

func (this *Sql) Commit() (bool, string) {
	if this.tx == nil {
		//slimSqlLog("Tx", "Commit tx failed because tx is null.")
		return false, "Tx is null!"
	} else {
		var err error
		err = this.tx.Commit()
		this.tx = nil
		if err != nil {
			slimSqlLog("Tx", "Commit tx failed because "+err.Error())
			return false, err.Error()
		} else {
			//slimSqlLog("Tx", "Commit tx success.")
			return true, ""
		}
	}
}

func (this *Sql) Rollback() (bool, string) {
	if this.tx == nil {
		//slimSqlLog("Tx", "Rollback tx failed because tx is null.")
		return false, "Tx is null!"
	} else {
		var err error
		err = this.tx.Rollback()
		this.tx = nil
		if err != nil {
			slimSqlLog("Tx", "Rollback tx failed because "+err.Error())
			return false, err.Error()
		} else {
			//slimSqlLog("Tx", "Rollback tx success.")
			return true, ""
		}
	}
}

/**
 * Lock table
 */
func (this *Sql) Lock(tables string, wirte bool) (bool, string) {
	tablesStr := this.safeInSql(tables, 2)
	wirteorread := "READ"
	if wirte == true {
		wirteorread = "WRITE"
	}
	sqlstr := "LOCK TABLE " + tablesStr + " " + wirteorread
	_, err := this.getDbW().Exec(sqlstr)
	if err != nil {
		slimSqlLog("Lock", "Lock "+tablesStr+" failed because "+err.Error())
		return false, err.Error()
	}
	//slimSqlLog("Lock", "Lock "+tablesStr+" success")
	return true, ""
}

/**
 * Unlock
 */
func (this *Sql) Unlock() (bool, string) {
	sqlstr := "UNLOCK TABLES"
	_, err := this.getDbW().Exec(sqlstr)
	if err != nil {
		slimSqlLog("UnLock", "UnLock failed because "+err.Error())
		return false, err.Error()
	}
	//slimSqlLog("UnLock", "UnLock success")
	return true, ""
}

/**
 * Lock row / for update
 */
func (this *Sql) LockRow() *Sql {
	if this.tx == nil {
		//slimSqlLog("LockRow", "Lockrow failed because tx is null.")
		return this
	} else {
		this.forupdate = true
		return this
	}
}

func (this *Sql) Clear() *Sql {
	if this.tx != nil {
		this.Commit()
	}
	this.mustMaster = false
	this.fieldsSql = ""
	this.tableName = ""
	this.conditionSql = ""
	this.joinSql = ""
	this.groupSql = ""
	this.havingSql = ""
	this.orderSql = ""
	this.pageSql = ""
	this.tx = nil
	this.pkSql = ""
	this.forupdate = false
	return this
}

func slimSqlLog(thetype string, content string) {
	fmt.Println("[SLIMSQL] " + thetype + ": " + content)
}

func (this *Sql) Ping() (bool, string) {
	err := this.getDbW().Ping()
	if err != nil {
		return false, err.Error()
	} else {
		return true, ""
	}
}

func (this *Sql) NewCondition() map[string]interface{} {
	return make(map[string]interface{})
}
