package driver

import (
	"fmt"
	"unsafe"
)

type Connection struct {
	db unsafe.Pointer

	backend SQLiteBackend
	uri     string
}

func Open(uri string, backend SQLiteBackend) (*Connection, error) {
	c := &Connection{
		uri:     uri,
		backend: backend,
	}
	flags := c.backend.OpenReadWrite() | c.backend.OpenCreate() | c.backend.OpenNoMutex() | c.backend.OpenURI()
	cUri := c.backend.CharPtr(c.backend.StringData(c.uri + "\x00"))
	var db unsafe.Pointer
	if rc := c.backend.OpenV2(cUri, unsafe.Pointer(&db), flags, nil); rc != c.backend.ResultOk() {
		c.backend.CloseV2(db)
		return nil, c.resultCodeToError(rc)
	}
	c.db = db
	c.backend.ExtendedResultCodes(db, 1)
	return c, nil
}

func (c *Connection) Close() error {
	if db := c.db; db != nil {
		c.db = nil
		if rc := c.backend.CloseV2(db); rc != c.backend.ResultOk() {
			return c.resultCodeToError(rc)
		}
	}
	return nil
}

func (c *Connection) PrepareStatement(sql string) (s *Statement, err error) {
	zSQL := sql + "\x00"
	cZQL := c.backend.CharPtr(c.backend.StringData(zSQL))
	var stmt unsafe.Pointer
	if rc := c.backend.PrepareV2(c.db, cZQL, unsafe.Pointer(&stmt)); rc != c.backend.ResultOk() {
		return nil, c.resultCodeToError(rc)
	}
	return &Statement{stmt: stmt, conn: c}, nil
}

func (c *Connection) Exec(sql string) error {
	sql += "\x00"
	cSQL := c.backend.CharPtr(c.backend.StringData(sql))
	if rc := c.backend.Exec(c.db, cSQL); rc != c.backend.ResultOk() {
		return c.resultCodeToError(rc)
	}
	return nil
}

type Statement struct {
	stmt unsafe.Pointer
	conn *Connection
}

func (s *Statement) SetInt(index int, value int64) error {
	if rc := s.conn.backend.BindInt64(s.stmt, index, value); rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

func (s *Statement) SetFloat(index int, value float64) error {
	if rc := s.conn.backend.BindDouble(s.stmt, index, value); rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

func (s *Statement) SetText(index int, value string) error {
	cValue := s.conn.backend.CharPtr(s.conn.backend.StringData(value))
	if rc := s.conn.backend.BindText(s.stmt, index, cValue, len(value)); rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

func (s *Statement) SetNull(index int) error {
	if rc := s.conn.backend.BindNull(s.stmt, index); rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

func (s *Statement) Exec() error {
	for {
		rc := s.conn.backend.Step(s.stmt)
		if rc == s.conn.backend.ResultRow() {
			continue
		} else if rc == s.conn.backend.ResultDone() {
			break
		} else {
			s.conn.backend.Reset(s.stmt)
			return s.conn.resultCodeToError(rc)
		}
	}
	if rc := s.conn.backend.Reset(s.stmt); rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

func (s *Statement) Reset() error {
	if rc := s.conn.backend.Reset(s.stmt); rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

func (s *Statement) Query() *ResultSet {
	return &ResultSet{s: s}
}

func (s *Statement) Close() error {
	rc := s.conn.backend.Finalize(s.stmt)
	s.stmt = nil
	if rc != s.conn.backend.ResultOk() {
		return s.conn.resultCodeToError(rc)
	}
	return nil
}

type ResultSet struct {
	s *Statement
}

func (r *ResultSet) Next() (bool, error) {
	rc := r.s.conn.backend.Step(r.s.stmt)
	if rc == r.s.conn.backend.ResultRow() {
		return true, nil
	}
	if rc == r.s.conn.backend.ResultDone() {
		return false, nil
	}
	return false, r.s.conn.resultCodeToError(rc)
}

func (r *ResultSet) ColumnCount() int {
	return r.s.conn.backend.ColumnCount(r.s.stmt)
}

func (r *ResultSet) ColumnName(i int) string {
	return r.s.conn.backend.ColumnName(r.s.stmt, i)
}

func (r *ResultSet) ColumnType(i int) int {
	return r.s.conn.backend.ColumnType(r.s.stmt, i)
}

func (r *ResultSet) GetFloat64(i int) float64 {
	if r.s.conn.backend.ColumnType(r.s.stmt, i) == r.s.conn.backend.ResultNull() {
		return 0.0
	}
	return r.s.conn.backend.ColumnDouble(r.s.stmt, i)
}

func (r *ResultSet) GetInt64(i int) int64 {
	if r.s.conn.backend.ColumnType(r.s.stmt, i) == r.s.conn.backend.ResultNull() {
		return 0
	}
	return r.s.conn.backend.ColumnInt64(r.s.stmt, i)
}

func (r *ResultSet) GetText(i int) (val string) {
	if r.s.conn.backend.ColumnType(r.s.stmt, i) == r.s.conn.backend.ResultNull() {
		return ""
	}
	return r.s.conn.backend.ColumnText(r.s.stmt, i)
}

func (r *ResultSet) Close() error {
	conn := r.s.conn
	stmt := r.s.stmt
	rc := conn.backend.Reset(stmt)
	r.s = nil
	if rc != conn.backend.ResultOk() {
		return conn.resultCodeToError(rc)
	}
	return nil
}

type Error struct {
	rc  int
	msg string
}

func (err *Error) Error() string {
	return fmt.Sprintf("sqlite3: %s [%d]", err.msg, err.rc)
}

var panicOnError = false

func PanicOnError() {
	panicOnError = true
}

func (c *Connection) resultCodeToError(rc int) error {
	var err error
	if c.db != nil {
		err = &Error{rc, c.backend.ErrMsg(c.db)}
	} else {
		err = &Error{rc, c.backend.ErrStr(rc)}
	}
	if panicOnError {
		panic(err)
	}
	return err
}
