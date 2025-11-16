package driver

import "unsafe"

type SQLiteBackend interface {
	OpenReadWrite() int
	OpenReadOnly() int
	OpenCreate() int
	OpenMemory() int
	OpenNoMutex() int
	OpenFullMutex() int
	OpenURI() int
	OpenExtendedResultCode() int

	OpenV2(filename unsafe.Pointer, ppDb unsafe.Pointer, flags int, zVfs unsafe.Pointer) int
	CloseV2(db unsafe.Pointer) int
	Exec(db unsafe.Pointer, sql unsafe.Pointer) int

	PrepareV2(db unsafe.Pointer, zSql unsafe.Pointer, ppStmt unsafe.Pointer) int
	Step(stmt unsafe.Pointer) int
	Reset(stmt unsafe.Pointer) int
	Finalize(stmt unsafe.Pointer) int

	BindInt64(stmt unsafe.Pointer, index int, value int64) int
	BindDouble(stmt unsafe.Pointer, index int, value float64) int
	BindText(stmt unsafe.Pointer, index int, value unsafe.Pointer, n int) int
	BindNull(stmt unsafe.Pointer, index int) int

	ColumnCount(stmt unsafe.Pointer) int
	ColumnName(stmt unsafe.Pointer, i int) string
	ColumnType(stmt unsafe.Pointer, i int) int
	ColumnDouble(stmt unsafe.Pointer, i int) float64
	ColumnInt64(stmt unsafe.Pointer, i int) int64
	ColumnText(stmt unsafe.Pointer, i int) string
	ColumnBytes(stmt unsafe.Pointer, i int) int

	ErrMsg(db unsafe.Pointer) string
	ErrStr(rc int) string

	CharPtr(p unsafe.Pointer) unsafe.Pointer
	StringData(s string) unsafe.Pointer

	ResultOk() int
	ResultRow() int
	ResultDone() int
	ResultNull() int
}
