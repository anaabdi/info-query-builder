# INFO-QUERY-BUILDER

help to generate a query for specific project

### Prerequisites


### Installing

go install

info-query-builder -h

Usage of info-query-builder:
  -port string
        openend port for serving request (default "3222")
  -prodURL string
        base url for accessing images in production (default "http://localhost/images")
  -stgURL string
        base url for accessing images in staging (default "http://localhost/images")

### Testing

Run it
```
$ info-query-builder -port 3222 -prodURL https://testing.com -stgURL https://testing.com
```

Note:
```
if you want to apply query for available cities, then just send an empty cities value.
like below.
"cities": []

if you want to specify for certain cities only, then fill it like below:
"cities": ["Jakarta","Bandung"]

for now, "prev_image_url" filled by filename of previous image that is inserted to your system, and that field is required for generating update query.
```


Generate query insert
```
curl -X POST \
  http://localhost:3222/query/generate \
  -H 'Content-Type: application/json' \
  -d '{
	"info_type": "promo",
	"title": "Promo Good life",
	"message": "Get 50% discount for living a good life",
	"start_time": "2019-08-16",
	"end_date": "2019-08-18",
	"promocode": "GOODLIFE",
	"cities": []
}'
```

Generate query update
```
curl -X POST \
  http://localhost:3222/query/update \
  -H 'Content-Type: application/json' \
  -d '{
	"info_type": "promo",
	"title": "Promo Real Life",
	"message": "Get 50% discount for always wake up in 5AM",
	"start_time": "2019-08-01",
	"end_date": "2019-09-30",
	"created_at": "2019-08-01",
	"promocode": "WAKEUP5",
	"prev_image_url": "WAKEUP5.png",
	"cities": ["Jakarta","Bandung"]
}'
```

### Contributing

Feel free to make this better