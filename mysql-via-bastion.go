package main

import (
	"database/sql"
	"fmt"
	_ "github.com/Go-SQL-Driver/MySQL"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
)

// Constants for the program, including location of PEM file,
// database server name, bastion server name, etc...
const PEM_FILE = "/DIR/TO/YOUR/secret.pem"
const BASTION_USER = "BASTION-SERVER-USERNAME"
const BASTION_SERVER = "SOME.BASTION-BOX.COM:22"
const DB_SERVER = "RDS.SERVER.NAME:3306"
const DB_DSN = "USER:PASSWORD@tcp(localhost:3306)/DATABASE"

func GetPublicKeyFileFrom(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Drat, couldn't read the pem file")
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Println("Drat, couldn't parse the buffer")
		return nil
	}

	return ssh.PublicKeys(key)
}

func main() {

	config := &ssh.ClientConfig{
		User: BASTION_USER,
		Auth: []ssh.AuthMethod{
			GetPublicKeyFileFrom(PEM_FILE),
		},
	}

	// Connect to the bastion server
	conn, err := ssh.Dial("tcp", BASTION_SERVER, config)
	if err != nil {
		log.Fatalf("Unable to connect to bastion server: %s", err)
	}
	defer conn.Close()

	fmt.Println("Connected to the bastion server, now to set up the database connection...")

	// Set up the connection to the remote database server
	remote, err := conn.Dial("tcp", DB_SERVER)
	if err != nil {
		log.Fatalf("Unable to connect to DB server: %s", err)
	}
	defer remote.Close()

	// local connection - figured not to make this a constant to confuse things
	local, err := net.Listen("tcp", "localhost:3306")
	if err != nil {
		log.Fatalf("Unable to connect to localhost: %s", err)
	}
	defer local.Close()

	go func() {
		for {

			l, err := local.Accept()
			if err != nil {
				log.Fatalf("listen Accept failed %s", err)
			}

			go func() {
				_, err := io.Copy(l, remote)
				if err != nil {
					log.Fatalf("io.Copy failed: %v", err)
				}
			}()

			go func() {
				_, err := io.Copy(remote, l)
				if err != nil {
					log.Fatalf("io.Copy failed: %v", err)
				}
			}()

		}
	}()

	// Note the use of localhost in the dsn! We are connecting to localhost on port 3306
	// and that is being passed over to the remote server
	db, err := sql.Open("mysql", DB_DSN)

	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

    //
	// Whee, we should now be connected and from here on out it's MySQL-stuff...
    //

	var count uint64
	row := db.QueryRow("SELECT COUNT(*) FROM SOME_TABLE")
	err = row.Scan(&count)

	if err != nil {
		log.Fatalf("SQL error: %s", err)
	}

	fmt.Printf("total row count from table is %d", count)

	// This isn't necessary, but an example of an anonymous
	// function that is invoked by passing the db parameter
	// below. If the function had been prepended by 'go', that
	// would make it a goroutine, which means it would run
	// asynchronously. The problem is that the whole program
	// will end before it's done, so it never gets run.
	func(db *sql.DB) {

		fmt.Println("\nOkay starting with rows...")
		rows, err := db.Query("Select SOME_NUMBER, SOME_STRING from SOME_TABLE")
		if err != nil {
			log.Fatalf("Drat, couldn't execute the query")
		}

		// Now go through all the rows...
		for rows.Next() {
			var someNum int
			var someString string
			// And populate the variables with the data...
			err := rows.Scan(&someNum, &someString)
			if err != nil {
				log.Fatalf("Hmm, couldn't get a row")
			}

			fmt.Printf("%d - %s\n", someNum, someString)
		}
	}(db) // Invoke the anonymous function with the database connection
}
