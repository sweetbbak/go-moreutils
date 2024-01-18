# Cut
implements --delimiter and --fields of `cut`
an improvement is that this version of the tool accepts
multiple characters as a delimiter.

```sh
echo -e '1||2||3\n4 5 6\n7||8 9' | cut -d '||' -f2
# outputs
1
7
```

```sh
# if -d isn't specified it will infer the delimiter to be a space
echo -e '1 2 3\n4 5 6\n7 8 9' | cut -f1
# outputs
1
4
7
# examples
cat example.txt | cut -d ' ' -f10
# print a range of fields
cat example.txt | cut -d ' ' -f 9-33
# print from the 5th field onwards
cat example.txt | cut -d ' ' -f 5-
# print from beginning to five fields
cat example.txt | cut -d ' ' -f -5
# print 1 2 3 4 5 6 and 11-19
cat example.txt | cut -d ' ' -f 1,2,3,4,6,11-19
# overlapping options will only print once (note that this is non-sensical)
cat example.txt | cut -d ' ' -f +10,1-5,9-
```
