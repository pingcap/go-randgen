/*

test embed lua

used in unittest, if you change it, you should change the relational unittest in `cmd/go-randgen/gentest_test.go`

*/

{
f={a=1, b=3}
arr={0,2,3,4}
}

query:
  {print(arr[f.a])} | {print(arr[f.b])}