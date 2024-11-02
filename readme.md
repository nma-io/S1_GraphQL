# S1 GraphQL Tool

I've been experimenting with the new GraphQL capability on SentinelOne. While powerfull, there is very limited documentation online.

I'm building this repository to help others get started. This will never be a fully functional tool, but merely a pathway to perform specific actions.

Please open a PR if you want to add any functionality.

As of now, I'm hardcoding the endpoint/region - you will need to modify constants.go with your endpoint address.

The API key pulls from environment variables, so please update your local environment and store the token in the value "S1_TOKEN"


You'll need to fetch your scopeid by visiting Policy & Settings in Singularity Operations Center - then select "Scope Info" midway down the page.


You could quickly close alerts with:

go run . --scope 1234567890987654321 -c -start 2024-01-01 -end 2024-05-01 -product Identity -m "Training data - Closing" -fp

To prevent a runaway process, I am currently not paging through the output - its handling a maximum of 1000 events per run.


Useful files but not used by the code:


> schema.query can be used to pull the full GraphQL Schema - I am using this with Altair to learn the mappings.

> schema.response is the json response at the time I started working with this - might be helpful if you just need to do a quick lookup.

Version 2024.0.1 Snyk output:

```

Testing S1_GraphQL ...


✔ Test completed

Organization:      nalbright
Test type:         Static code analysis
Project path:      code/S1_GraphQL

Summary:

✔ Awesome! No issues were found.

```


Original Author: Nicholas Albright (@nma-io)


Contributors:

None yet.
