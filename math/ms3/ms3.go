/*
package ms3 is a 32bit 3D math package based around the Vec type.

This is a different take compared to mgl32, a package that
treats all data equally by giving them their own namespace with
methods. This package is more similar to gonum's take on spatial operations
where package-level functions like `Add` are reserved specifically for vector
operations, which tend to be the most common. This aids in readability
since long string of operations using methods can be remarkably hard to follow.

The name roughly stands for (m)ath for (s)hort floats in (3)D.
"short" since there are no native 16 bit floats in Go.
*/
package ms3
