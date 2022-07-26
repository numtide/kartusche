# Code
Code contained in Kartusche is JavaScript.

If you're a seasoned JS developer, don't get too excited - the code is not run by Node or anything similar.
In fact, the code is executed by [goja](https://github.com/dop251/goja) a JavaScript implementation written in Golang.
Reason for using JS is to have a language that can be succinct and powerful enough to write the 'glue' code on top of primitives written in Go.

On the other hand, you can profit of existing JS knowledge, since [goja](https://github.com/dop251/goja) is a full fledged ES 5.1 supports almost all of [ES6](https://262.ecma-international.org/6.0/).

This has following consequences:

* Only 'off the shelf' ECMAScript 5.1 (strings, arrays, Date, JSON, ...) with addition of Kartusche specific functions are available.
* Kartusche does not support [npm](https://www.npmjs.com/) packages.
* Executed code is always synchronous. 
Promises and other async constructs such as async/await or [rxjs](https://rxjs.dev/) are not supported.

