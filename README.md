# Pun Threads

This is a command-line go program that scrapes reddit posts, parses comment threads, and sends them to the ChatGPT api to evaluate for whether they qualify as "puns".


A command-line argument specifies which function to execute:

- gather: Fetch new posts in json files from reddit and save the results to local raw files
- rate: walk the collection of local reddit content files, locate "well-shaped" comment threads and send them to chatgpt to get a single pun rating for the comment thread, on a 0 to 10 scale. Chatgpt's responses are saved to a local mongo database
- review: walk the mongo database and print to stdout the threads that have a sufficiently high rating, e.g. 8/10

A ChatGPT API key is required to run the rate function. 

Some output examples: https://docs.google.com/document/d/1pOsA2EmLBtGKc7UDEaORMlAiWEXbr8l3WveHnVUvf_Y/
