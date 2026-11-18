TOPIC 1: DURABILITY AND ATOMICITY

1. What is a database?
Today everyone calls anything a database. A lot of marketing campaigns are just calling anything a database, but this is not right. Someone would literally say excel sheets, cache servers are DB, while they are not entirely wrong, we need to find out what a real traditional DB was. It had some featires that made it stand out.
- They persisted data to disk
- They are disk-based - that means that they can work with larger-than-memory data.
- They are usually implemented from scratch not as a wrapper over other datababes. 

Exmaples of DB that posses this traits are MySQL, Postgres, SQLite, etc

2. A DB is more than just a data format
Persisiting data to disk(durability) is the number one criterion of traditional DB. This is why mobile phones use SQLite, a file-based DB. But why not use file if the DB is just a file anyway? Because persisting has some concrete requirements. At some point, the data added to the DB is guaranteed to persist even if the machine crashes. The two main requirements for persisting are: 
- To survive crashes
- To confirm the state of durability
