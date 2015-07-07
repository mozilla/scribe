all:

runtests: test.json
	$(SCRIBECMD) -e -f test.json

test.json: test-template.json
	cat test-template.json | sed 's,<<t>>,$(shell pwd),' > test.json

clean:
	rm -f test.json
