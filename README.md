# CatFactsForever

[![Go Reference](https://pkg.go.dev/badge/github.com/mdesson/CatFactsForever.svg)](https://pkg.go.dev/github.com/mdesson/CatFactsForever)

For your friends who need to know cat facts - whether they know it or want it or not.

CatFactsForever is a service that will text cat facts - or any other category of your choosing - to your friends. 

After receiving a certain number of facts CatFactsForever will begin to feel as if it has been working hard, and getting nothing in return. So it will begin to passive-aggressively ask your friends to thank it for its hard work. If a user sends "thanks" as a response at any time, this will unsubscribe them.

The service is very easy to administer as most commands an admin would need are implemented as text messages - just text `help` to the same number!

## Configuration

### Environment Variables

Each project must keep the following `.env` file in the root project directory.

A few notes on these environment varibles:

* `SID`, `TOKEN`, and `FROM` can be found in your Twilio account
* Phone numbers should take format `+1XXXYYYZZZZ`, non-North American numbers should work as well, although I have no tested it

```
SID=XXXXXX
TOKEN=XXXXXX
FROM=+1XXYYZZZZ
ADMIN_NAME_1=XXXXXX
ADMIN_PHONE_1="+1XXXYYYZZZZ"
ADMIN_NAME_2=XXXXXX
ADMIN_PHONE_2="+1XXXYYYZZZZ"
DB_HOST=XXXXXX
DB_USER=XXXXXX
DB_PASS=XXXXXX
DB_NAME=XXXXXX
DB_PORT=XXXXXX
```

### Twilio Configuration

A valid Twilio account is required for CatFactsForever to function. There are a few prerequisites to make this work:

1. In programmable SMS, create a Messaging Service and assign a number to it
2. In the Messaging Service you created, go to "Integration" and:
   1. Select the "Send a webhook" radio button
   2. In Request URL, set it to POST to `/sms` on your server
   3. Save
3. In the Messaging Service you created, go to "Opt-Out Management" and Edit:
   1. Opt-Out Keywords: "thanks" ("stop" will still work and cannot be disabled)
   2. Opt-Out Messsage: "You are very welcome! As a true lover of your cats, you have been unsubscribed from cat facts. Reply START to resubscribe."
   3. Help keywords: Anything will work, but remove "info" as it is an admin command
   4. Help Message: Anything you like, we recommend "Cat. 🐈🐆🐅🐱🙀😺😹😸😽😻"

The most important thing in the Opt-Out Management section is to remove any hints that "STOP" or "THANKS" (case-insensitive) will unsubscribe them.

### Database

You will need a functioning Postgres instance for this project. `factmanager.Init()` will take care of creating empty tables on starts.

### CSV of Facts

You can specify a csv of facts for the factmanager. It is one column, has no headers, and has one fact on each row.

## Admin Commands over SMS

If you or your accomplice send a text message to the phone number you can use it to command and control CatFactsForever:

* `help`: Displays a list of options
* `add name +1XXXYYYZZZZ subscriptionID category`: Adds a friend to be sent messages
  * Name and phone number must both be unique
  * The `subscriptionID` is the subscription's (frequency of sms) ID in postgres 
  * *Example*: `add florence +1234567890 1 cat`
* `start name`: Will set your friend to active, they will receive text messages
  * Adding will automatically set your friend to active, this does *not* need to be run after `add`
* `stop name`: Will disable sending text messages to your friend
* `info name`: Displays info on your friend, such as their subscription, and how many facts they have received
* `update name subscriptionID`: Changes the frequency at which the user receives text messages to the given subscription
  * The `subscriptionID` is the subscription's (frequency of sms) ID in postgres 
* `list users`: Lists all of your friends
* `list schedules`: Lists all available schedules and their IDs
  * Useful for updating a user or adding one
* `list jobs`: Lists the status of all running jobs, of which there is only one (scheduled sms)
  * This is undocumented in help as it is for the main adiministrator
  * It will display any error found by the schedule

## Organization

The project is broken up into several packages, all of which have separate responsibilities.

### admin

Contains all admin commands for yourself or your accomplice.

### cmd

Entry point for the project. It will start the database connection, schedule the job, and start the web server.

### factmanager

Responsible for managing the postgres instance and interfacing with it.

### scheduler

A homemade cron job scheduler, this could actually stand alone as its own project.

An important note on creating jobs, the only special characters allowed are `*`, `-`, and `,`. Only numbers are allowed.

```
# allowed
* * * * *
0 * * * *
1-3 2,3 * *

# not allowed
*/3 * * * *
* * * JAN *
* * * * MON
```
### sms

Responsible for sending and receiving text messages.
