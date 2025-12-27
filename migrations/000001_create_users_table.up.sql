CREATE TABLE IF NOT EXISTS public.users (
                                            id SERIAL PRIMARY KEY,
                                            username TEXT UNIQUE NOT NULL,
                                            password TEXT NOT NULL,
                                            age INTEGER NOT NULL
);
