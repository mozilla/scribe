all:

runtests: test.json
	$(SCRIBECMD) -f test.json

test.json: test.json.template
	cat test.json.template | sed 's,<<t>>,$(shell pwd),' > test.json

clean:
	rm -f test.json
