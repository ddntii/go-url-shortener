# go-url-shortener

a simple url shortener for the command line. takes long urls and makes them short with tracking.

## what it does

- makes short codes for long urls
- lets you pick custom codes  
- counts how many times people click
- grabs page titles automatically
- won't make duplicates
- cleans up old unused links
- saves everything to a json file

## setup

```bash
git clone https://github.com/ddntii/go-url-shortener
cd go-url-shortener
go build -o urlsh
```

## how to use

### make a short url
```bash
# random code
./urlsh s https://example.com
# gives you: abc4

# your own code
./urlsh s https://example.com mylink
# gives you: mylink
```

### get the original url back
```bash
./urlsh e abc4
# shows: https://example.com
#        Title: Example Domain  
#        Clicks: 1
```

### see all your urls
```bash
./urlsh l
# shows everything you've saved
```

### check your stats
```bash
./urlsh stats  
# shows total urls, clicks, most popular, etc
```

### delete old stuff
```bash
./urlsh clean
# removes urls with 0 clicks that are 30+ days old
```

### remove a specific url
```bash
./urlsh delete abc4
# deletes that short code
```

## all the commands

- `urlsh s <url>` - make short url (can also use `shorten`)
- `urlsh s <url> <code>` - make short url with your code
- `urlsh e <code>` - expand short url (can also use `expand`) 
- `urlsh l` - list everything (can also use `list`)
- `urlsh stats` - show statistics
- `urlsh clean` - remove old unused urls
- `urlsh delete <code>` - delete specific url (can also use `del` or `rm`)

## where stuff gets saved

everything goes in `urls.json` and looks like this:
```json
{
  "items": {
    "abc4": {
      "url": "https://example.com",
      "created_at": "2025-01-15T10:30:00Z",
      "clicks": 5,
      "title": "Example Domain",
      "last_click": "2025-01-16T14:22:00Z"
    }
  }
}
```

## how the codes work

- uses random letters and numbers (no confusing ones like 0, O, 1, l)
- starts with 4 character codes
- makes them longer if you have lots of urls
- won't give you the same code twice

## what you need

- go 1.19 or newer



## thats it enjoy it ! 
- that's it, no other stuff needed
