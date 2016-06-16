# gcc compiler to use
CC=gcc

# compiler options
CFLAGS=-o
#CFLAGS=-c -Wall
CFLAGS2=-lm -Wall

src = main current cos frequenz power value voltage smartpi_autostart all_rrd

all: $(src)
	

main: all.c 
	$(CC) $(CFLAGS) main all.c

current: current.c
	$(CC) $(CFLAGS) current current.c

cos: cos.c
	$(CC) $(CFLAGS) cos cos.c

frequenz: frequenz.c
	$(CC) $(CFLAGS) frequenz frequenz.c

power: power.c
	$(CC) $(CFLAGS) power power.c

value: value.c
	$(CC) $(CFLAGS) value value.c

voltage: voltage.c
	$(CC) $(CFLAGS) voltage voltage.c

smartpi_autostart: smartpi_autostart.c
	$(CC) $(CFLAGS) smartpi_autostart smartpi_autostart.c
	
all_rrd: all_rrd.c
	$(CC) $(CFLAGS) all_rrd all_rrd.c $(CFLAGS2)
	
install:
	echo make install to be implemented...
	
clean:
	rm -f *o $(src)