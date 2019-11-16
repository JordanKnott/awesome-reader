lines = open("README.md").readlines()

lines = [l.strip() for l in lines]



def get_level(line):
    level = 0
    for c in line:
        if c != '#':
            break
        level += 1
    return level


sections = {}
current_section = None
for line in lines:
    level = 0
    if line.startswith("#"):
        level = get_level(line)
        print("{}: {}".format(level, line))
        sections = {
