This package enables PGX to use parameters of type go-utility/v2/uuid. To enable support, register the type in the `AfterConnect` callback before connecting to the database.

```golang
dbconfig, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
	// handle error
}
dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
	conn.ConnInfo().RegisterDataType(pgtype.DataType{
		Value: &pgxcompat.UUID{},
		Name:  "uuid",
		OID:   pgtype.UUIDOID,
	})
	return nil
}```
