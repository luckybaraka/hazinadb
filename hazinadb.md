Storage engines are not easy to implement. They need more than just what we think a DB is. 
They are built on a few strong principles.

1. Atomocity and Durability
- From your ACID learnings I am sure you have encountered this. 
- ACID spans from a larger topic called Transactions. And transaction is away of grouping work(reads, or writes or both reads and writes) into one logical unit of work so that they can be executed together. We will look deeper into this later, but the thing around this is the fact that the stoarge engine has to think around persisiting data with fsync() and then it also has to think about the worst cases, how will crash recovery do happen? 

2. KV Store based on B-Trees. 
This is just thinking around disk-based data structures and space management with a free list

3. Relational DB on top of KV.
This one here is How tables and indexes are mapped to lowe-level B-trees.
And SQL-like querry control for transactions. 

4. Concurrency control for transactions. What happens if two or more processes are trying to access the same variable? 


