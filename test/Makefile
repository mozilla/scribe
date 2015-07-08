TESTDIRS = filecontent filename evrtest

all:

runtests:
	for x in $(TESTDIRS); do \
		$(MAKE) -C $$x runtests; \
	done

clean:
	for x in $(TESTDIRS); do \
		$(MAKE) -C $$x clean; \
	done
