# Ticket Service

## Description
This is a simple ticket service that facilitates the discovery, temporary hold, and final reservation of seats within a high-demand performance venue.

The following three routes need to be built to enable allocating  tickets to multiple purchases.
The solution needs to ensure that the allocation does not drop  below 0, and the purchased amounts are not greater than the  allocation given.
Taking payment is out of scope for this problem.

Swagger documentation is available at `http://localhost:8080/swagger`

## Technologies
- golang
- chi router
- postgres
- docker

## Setup
1. Clone the repository
2. Run `docker-compose up` to start the postgres database and the application
3. The application will be available at `http://localhost:8080`
4. The swagger documentation is available at `http://localhost:8080/swagger`
5. The database is available at `localhost:5432` with the username `postgres` and password `password`

## Database
The database has two tables:
1. **ticket_options**
    - `id` - id of the ticket option
    - `name` - name of the ticket option
    - `description` - description of the ticket option
    - `allocation` - Number of seats to sell
2. **purchases** (to store the purchases made by the users)
    - `id` - id of the purchase
    - `ticket_option_id` - id of the ticket option
    - `quantity` - Number of tickets purchased
    - `user_id` - id of the user

## Routes
1. **POST /tickets_option**
    - Request:
        - `name` - name of the ticket option
        - `description` - description of the ticket option
        - `allocation` - Number of seats to sell
    
    - Response:
        - `id` - id of the ticket option
        - `name` - name of the ticket option
        - `description` - description of the ticket option
        - `allocation` - Number of seats to sell
        
    - Description:
        - This route is used to create a ticket option with the given name, description, and allocation.
2. **GET /tickets_option/{id}**
    - Response:
        - One of ticket options
            - `id` - id of the ticket option
            - `name` - name of the ticket option
            - `description` - description of the ticket option
            - `allocation` - Number of seats to sell
            
    - Description:
        - This route is used to get all the ticket options.
3. **POST /tickets_option/{id}/purchases**
    - Request:
        - `quantity` - Number of tickets to purchase
        - `user_id` - id of the user
    
    - Response:
        - Status code 200 if the purchase is successful
        - Status code 400 if the purchase is unsuccessful
        
    - Description:
        - This route is used to purchase tickets for the given ticket option. The quantity of tickets purchased should not exceed the allocation of the ticket option.
