TESTDIRS = filecontent

all:

runtests:
	for x in $(TESTDIRS); do \
		cd $$x && $(MAKE) runtests; \
	done

clean:
	for x in $(TESTDIRS); do \
		cd $$x && $(MAKE) clean; \
	done
