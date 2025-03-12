const database = 'lm';
const collection = 'tracks';

// Create a new database.
use(database);

// Create a new collection.
db.createCollection(collection);
