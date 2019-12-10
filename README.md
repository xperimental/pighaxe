# piGHaxe - A Pickaxe Tool for Github

This is a small command-line tool that can be used to search for a regular expression in multiple Github repositories.

**Note:** This is still an early prototype.

## Usage

```plain
Usage of pighaxe:
  -h, --host string             GitHub host to use. (default "github.com")
      --http-timeout duration   Timeout for HTTP Requests. (default 5s)
  -v, --log-level string        Logging level. (default "info")
  -o, --organization string     Limit search to certain organization.
```

The search term is interpreted as a regular expression. If it does not contain any capturing groups, then the command will produce CSV output with three fields:

- Repository URL
- Filename
- Matched line

If the regular expression contains capturing groups then the output will contain the groups instead of the whole matched line.

For example, this can be used to search through the reporitories of an organization and check which version of a certain library they are using:

```bash
$ pighaxe -o example 'library-(.+)'
repo,file,group0
https://github.com/example/project1,dependencies.txt,1.0.0
https://github.com/example/project2,dependencies.txt,1.1.0
``` 
