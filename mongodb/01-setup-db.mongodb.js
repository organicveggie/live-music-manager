const database = 'lm';
const collection = 'live-music';

// Create a new database.
use(database);

// Create a new collection.
db.createCollection(collection);
