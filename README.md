# Autorest
A simple server that automatically creates a RESTful interface to a database. Currently only MySQL is supported. Pull requests are welcome to add support for other databases.

## How It Works
Autorest will connect to a database, parse the schema and start a server that will create a RESTful interface. It builds queries in a secure manner to prevent SQL injection. Each table will, by default, correspond to one endpoint. For example, if our database has two tables, `users` (columns id, first_name, last_name) and `products` (columns id, name), **autorest** will create the following endpoints:

- host:port/rest/users
- host:port/rest/products

Each endpoint supports the 4 main HTTP verbs, GET, POST, PUT, and DELETE. JSON is expected for the request body of POST and PUT requests. The specific HTTP calls that **autorest** will then respond to are the following:

- GET host:port/rest/users - Gets all users
- GET host:port/rest/users/:id - Get a single user
- POST host:port/rest/users - Create a new user
- PUT host:port/rest/users/:id - Update a user
- DELETE host:port/rest/users/:id - Delete a user

## Examples
### Setup the Server
```
func main() {
  credentials := autorest.DatabaseCredentials{
    Username: "root",
    Password: "admin",
    Host: "localhost",
    Name: "my_db",
    Port: "3306"
  }
  server := autorest.NewServer(credentials)
  server.Run("80")
}
```
