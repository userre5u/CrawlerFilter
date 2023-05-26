+++++++++++++++++++++++++++++++++++++++++++
|   Simple Golang Crawler-Filter V1       |
+++++++++++++++++++++++++++++++++++++++++++


The project contains two main components:
-----------------------------------------------
1) A lambda function that listens for requests, saves the request's metadata in S3 object and redirects the device based on request's criteria
    a) if the request is valid (not crawler) -> redirect to a nice template
    b) if the request is invalid (crawler) -> redirect to bad template
2) A client that downloads the objects from S3 bucket, parse the content and save it in a mysql database

Request is valid if [A-G] are passed:
    [A] -> IP is not blacklisted
    [B] -> Country is valid
    [C] -> Method is valid
    [D] -> Path is valid
    [E] -> User Agent is mobile (according to regex)
    [F] -> Session Key is valid
    [G] -> Body does not contain suspicious keywords