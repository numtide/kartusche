expect = {
    equal: (current, expected) => {
        if (current !== expected) {
            throw new Error(`expected ${current} to equal to ${expected}`)
        }
    }
}