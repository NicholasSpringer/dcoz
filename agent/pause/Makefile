CC = gcc

EXE = dcoz-pause

CFLAGS = -Wall -pipe 

SOURCEDIR = src
BUILDDIR = build

SOURCES = $(wildcard $(SOURCEDIR)/*.c)
OBJECTS = $(patsubst $(SOURCEDIR)/%.c,$(BUILDDIR)/%.o,$(SOURCES))

all: project

project: $(OBJECTS)
	$(CC) $(OBJECTS) $(CFLAGS) -o $(EXE)

$(OBJECTS): $(BUILDDIR)/%.o : $(SOURCEDIR)/%.c
	@mkdir -p $(BUILDDIR)
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -rf $(BUILDDIR) $(EXE)

rebuild: clean all

.PHONY : clean
.SILENT : clean