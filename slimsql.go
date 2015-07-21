package slimmysql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

var sqlDB *sql.DB
var safeMode bool
var prefix string

/**
 * Init sql conn
 */
func InitSql(user string, pass string, ip string, port string, db string, pre string, safe bool) error {
	slimSqlLog("Init", "user:"+user+" pass:"+pass+" ip:"+ip+" port:"+port+" db:"+db)
	var err error
	sqlDB, err = sql.Open("mysql", user+":"+pass+"@tcp("+ip+":"+port+")/"+db+"?charset=utf8")
	sqlDB.SetMaxOpenConns(2000)
	sqlDB.SetMaxIdleConns(1000)
	sqlDB.Ping()
	prefix = pre
	safeMode = safe
	return err
}

type Sql struct {
	fieldsSql    string
	tableName    string
	conditionSql string //where
	joinSql      string
	groupSql     string
	havingSql    string
	orderSql     string
	pageSql      string
	tx           *sql.Tx //Transaction
	pkSql        string  //primary key
}

/**
 * Set table name,use prefix
 */
func (this *Sql) Table(tablename string) *Sql {
	this.tableName = prefix + safeInSql(tablename, 1)
	return this
}

/**
 * Set table name,no prefix
 */
func (this *Sql) TrueTable(tablename string) *Sql {
	this.tableName = safeInSql(tablename, 1)
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
	this.fieldsSql = safeInSql(filed, 1)
	return this
}

/**
 * Join table
 * example:"INNER JOIN B on A.id = B.aid"
 */
func (this *Sql) Join(joinsql string) *Sql {
	this.joinSql = safeInSql(joinsql, 2)
	return this
}

/**
 * group by
 */
func (this *Sql) Group(g string) *Sql {
	this.groupSql = safeInSql(g, 1)
	return this
}

/**
 * having after group by
 * example:"age > 24"
 */
func (this *Sql) Having(h string) *Sql {
	this.havingSql = safeInSql(h, 2)
	return this
}

/**
 * order by
 * example:"id desc"
 */
func (this *Sql) Order(order string) *Sql {
	this.orderSql = safeInSql(order, 2)
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
		this.conditionSql = safeInSql(condition.(string), 2)
		break
	case map[string]interface{}:
		this.conditionSql = convertCondition(condition.(map[string]interface{}))
		break
	}
	return this
}

/**
 * convert condition to sql
 */
func convertCondition(condition map[string]interface{}) string {
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
			sql += join_sql + "(" + convertCondition(v.(map[string]interface{})) + ")"
			i++
			continue
		}
		temp := convertValue2String(v)
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
				sql += join_sql + temp_key + " > " + temp + " "
				break
			case "egt":
				sql += join_sql + temp_key + " >= " + temp + " "
				break
			case "lt":
				sql += join_sql + temp_key + " < " + temp + " "
				break
			case "elt":
				sql += join_sql + temp_key + " <= " + temp + " "
				break
			case "between":
				temp_v := strings.Split(temp, "|")
				sql += join_sql + " ( " + temp_key + " >= " + temp_v[0] + " AND " + temp_key + " <= " + temp_v[1] + " ) "
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

// /**
//  * convert condition map to sql str
//  */
// func convertCondition(condition map[string]interface{}, k string) string {
// 	//fmt.Println(condition)
// 	sql := ""
// 	if len(k) > 0 && k != "_" && condition["relation"] != nil {
// 		switch strings.ToUpper(condition["relation"].(string)) {
// 		case "GT":
// 			sql = k + " > " + convertValue2String(condition["value"])
// 			break
// 		case "EGT":
// 			sql = k + " >= " + convertValue2String(condition["value"])
// 			break
// 		case "LT":
// 			sql = k + " < " + convertValue2String(condition["value"])
// 			break
// 		case "ELT":
// 			sql = k + " <= " + convertValue2String(condition["value"])
// 			break
// 		case "BETWEEN":
// 			sql = k + " >= " + convertValue2String(condition["from"]) + " AND " + k + " <= " + convertValue2String(condition["to"])
// 			break
// 		case "IN":
// 			sql = k + " IN (" + convertValue2String(condition["value"]) + ")"
// 			break
// 		case "NOTIN":
// 			sql = k + " NOT IN (" + convertValue2String(condition["value"]) + ")"
// 			break
// 		case "LIKE":
// 			sql = k + " LIKE '%" + convertValue2String(condition["value"]) + "%'"
// 			break
// 		default:
// 			sql = " 1 = 1 "
// 			break
// 		}
// 		return sql
// 	}
// 	join := "AND"
// 	if condition["relation"] != nil && strings.ToUpper(condition["relation"].(string)) == "OR" {
// 		join = "OR"
// 	}
// 	i := 0
// 	join_sql := " "
// 	for k, v := range condition {
// 		if k == "relation" {
// 			continue
// 		}
// 		if i == 0 {
// 			join_sql = " "
// 		} else {
// 			join_sql = " " + join + " "
// 		}
// 		switch v.(type) {
// 		case string, int, int32, int64, float32, float64:
// 			temp := convertValue2String(v)
// 			if len(temp) > 0 {
// 				sql += join_sql + k + " = " + "'" + temp + "' "
// 			}
// 			i++
// 			break
// 		case map[string]interface{}:
// 			sql += join_sql + "(" + convertCondition(v.(map[string]interface{}), k) + ")"
// 			i++
// 			break
// 		}
// 	}
// 	return sql
// }

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
		sel_id := convertValue2String(id)
		if len(this.pkSql) > 0 {
			this.conditionSql = this.pkSql + " = '" + sel_id + "'"
		} else {
			this.conditionSql = "id = '" + sel_id + "'"
		}
	}
	sqlstr := this.GetSql(true) + " limit 1"
	slimSqlLog("Find", sqlstr)
	var rows *sql.Rows
	var err error
	if this.tx != nil {
		rows, err = this.tx.Query(sqlstr)
	} else {
		rows, err = sqlDB.Query(sqlstr)
	}
	if err != nil {
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
	slimSqlLog("Select", sqlstr)
	var rows *sql.Rows
	var err error
	if this.tx != nil {
		rows, err = this.tx.Query(sqlstr)
	} else {
		rows, err = sqlDB.Query(sqlstr)
	}
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
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
		this.fieldsSql = convertValue2String(filed)
	} else {
		this.fieldsSql = "*"
	}
	this.orderSql = ""
	this.pageSql = ""
	sqlstr := this.GetSql(false)
	slimSqlLog("Count", sqlstr)
	var rows *sql.Rows
	var err error
	if this.tx != nil {
		rows, err = this.tx.Query(sqlstr)
	} else {
		rows, err = sqlDB.Query(sqlstr)
	}
	if err != nil {
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
func safeInSql(sql string, level int) string {
	if safeMode {
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
func convertValue2String(v interface{}) string {
	var r string
	switch v.(type) {
	case string:
		r = safeInSql(v.(string), 1)
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
		tmp_v := convertValue2String(v)
		values = append(values, "\""+tmp_v+"\"")
	}
	sqlstr := " INSERT INTO `" + this.tableName + "` " + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.Join(values, ",") + ") "
	slimSqlLog("Insert", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = sqlDB.Exec(sqlstr)
	}
	if err != nil {
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
			setsql += ", " + k + "=\"" + convertValue2String(v) + "\" "
		} else {
			setsql += " " + k + "=\"" + convertValue2String(v) + "\" "
		}
		count++
	}
	if len(this.conditionSql) > 0 {
		//do not need pk

	} else {
		this.conditionSql = pk + " = \"" + convertValue2String(data[pk]) + "\" "
	}
	sqlstr += setsql + " WHERE " + this.conditionSql
	slimSqlLog("Update", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = sqlDB.Exec(sqlstr)
	}
	if err != nil {
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
	slimSqlLog("Delete", sqlstr)
	var r sql.Result
	var err error
	if this.tx != nil {
		r, err = this.tx.Exec(sqlstr)
	} else {
		r, err = sqlDB.Exec(sqlstr)
	}
	if err != nil {
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
		slimSqlLog("Tx", "Start new tx failed because tx is not null.")
		return false, "Tx not null!"
	} else {
		var err error
		this.tx, err = sqlDB.Begin()
		if err != nil {
			slimSqlLog("Tx", "Start new tx failed because "+err.Error())
			return false, err.Error()
		} else {
			slimSqlLog("Tx", "Start new tx success.")
			return true, ""
		}
	}
}

func (this *Sql) Commit() (bool, string) {
	if this.tx == nil {
		slimSqlLog("Tx", "Commit tx failed because tx is null.")
		return false, "Tx is null!"
	} else {
		var err error
		err = this.tx.Commit()
		this.tx = nil
		if err != nil {
			slimSqlLog("Tx", "Commit tx failed because "+err.Error())
			return false, err.Error()
		} else {
			slimSqlLog("Tx", "Commit tx success.")
			return true, ""
		}
	}
}

func (this *Sql) Rollback() (bool, string) {
	if this.tx == nil {
		slimSqlLog("Tx", "Rollback tx failed because tx is null.")
		return false, "Tx is null!"
	} else {
		var err error
		err = this.tx.Rollback()
		this.tx = nil
		if err != nil {
			slimSqlLog("Tx", "Rollback tx failed because "+err.Error())
			return false, err.Error()
		} else {
			slimSqlLog("Tx", "Rollback tx success.")
			return true, ""
		}
	}
}

func (this *Sql) Close() {
	sqlDB.Close()
}

func (this *Sql) Clear() *Sql {
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
	return this
}

func slimSqlLog(thetype string, content string) {
	fmt.Println("[SLIMSQL] " + thetype + ": " + content)
}
