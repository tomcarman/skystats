module skystats/core

go 1.22.6

// replace skystats/fetch => ../fetch
// require skystats/fetch v0.0.0-00010101000000-000000000000

require github.com/JamesLMilner/cheap-ruler-go v0.0.0-20191212211616-0919b75413a9

require (
	github.com/jackc/pgx/v5 v5.6.0
	github.com/sevlyar/go-daemon v0.1.6
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
