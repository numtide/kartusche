expect = {
    equal: (current, expected) => {
        if (current !== expected) {
            throw new Error(`expected ${current} to equal to ${expected}`)
        }
    },

    deepEqual: (current, expected) => {
        const currentJson = JSON.stringify(current)
        const expectedJson = JSON.stringify(expected)
        if (currentJson !== expectedJson) {
            throw new Error(`expected ${currentJson} to equal to ${expectedJson}`)
        }
    }
}