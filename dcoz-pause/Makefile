CC = gcc

EXE = dcoz-pause

CFLAGS = -Wall -pipe 

SOURCEDIR = src
BUILDDIR = build
BINDIR = bin

SOURCES = $(wildcard $(SOURCEDIR)/*.c)
OBJECTS = $(patsubst $(SOURCEDIR)/%.c,$(BUILDDIR)/%.o,$(SOURCES))

all: project

project: $(OBJECTS) 
	$(CC) $(OBJECTS) $(CFLAGS) -o $(BINDIR)/$(EXE)

$(OBJECTS): $(BUILDDIR)/%.o : $(SOURCEDIR)/%.c
	@mkdir -p $(BUILDDIR)
	@mkdir -p $(BINDIR)
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -rf $(BUILDDIR) $(BINDIR)

rebuild: clean all

.PHONY : clean
.SILENT : clean