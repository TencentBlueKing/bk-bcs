/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

const __whitespace = {
    ' ': true,
    '\t': true,
    '\n': true,
    '\f': true,
    '\r': true
}

const difflib = {
    defaultJunkFunction (c) {
        return __whitespace.hasOwnProperty(c)
    },

    stripLinebreaks (str) {
        return str.replace(/^[\n\r]*|[\n\r]*$/g, '')
    },

    stringAsLines (str) {
        const lfpos = str.indexOf('\n')
        const crpos = str.indexOf('\r')
        const linebreak = ((lfpos > -1 && crpos > -1) || crpos < 0) ? '\n' : '\r'

        const lines = str.split(linebreak)
        for (let i = 0; i < lines.length; i++) {
            lines[i] = difflib.stripLinebreaks(lines[i])
        }

        return lines
    },

    // iteration-based reduce implementation
    reduce (func, list, initial = 0) {
        let value = initial
        let idx = 0

        if (list && list.length) {
            value = list[0]
            idx = 1
            const len = list.length
            for (; idx < len; idx++) {
                value = func(value, list[idx])
            }
        }

        return value
    },

    // comparison function for sorting lists of numeric tuples
    ntuplecomp (a, b) {
        const mlen = Math.max(a.length, b.length)
        for (let i = 0; i < mlen; i++) {
            if (a[i] < b[i]) {
                return -1
            }
            if (a[i] > b[i]) {
                return 1
            }
        }

        return a.length === b.length ? 0 : (a.length < b.length ? -1 : 1)
    },

    calculate_ratio (matches, length) {
        return length ? 2.0 * matches / length : 1.0
    },

    // returns a function that returns true if a key passed to the returned function
    // is in the dict (js object) provided to this function; replaces being able to
    // carry around dict.has_key in python...
    isindict (dict) {
        return key => dict.hasOwnProperty(key)
    },

    // replacement for python's dict.get function -- need easy default values
    dictget (dict, key, defaultValue) {
        return dict.hasOwnProperty(key) ? dict[key] : defaultValue
    },

    SequenceMatcher (a, b, isjunk) {
        this.set_seqs = (a, b) => {
            this.set_seq1(a)
            this.set_seq2(b)
        }

        this.set_seq1 = a => {
            if (a === this.a) {
                return
            }
            this.a = a
            this.matching_blocks = this.opcodes = null
        }

        this.set_seq2 = b => {
            if (b === this.b) {
                return
            }
            this.b = b
            this.matching_blocks = this.opcodes = this.fullbcount = null
            this.chain_b()
        }

        this.chain_b = () => {
            const b = this.b
            const n = b.length
            const b2j = this.b2j = {}
            const populardict = {}
            for (let i = 0; i < b.length; i++) {
                const elt = b[i]
                if (b2j.hasOwnProperty(elt)) {
                    const indices = b2j[elt]
                    if (n >= 200 && indices.length * 100 > n) {
                        populardict[elt] = 1
                        delete b2j[elt]
                    } else {
                        indices.push(i)
                    }
                } else {
                    b2j[elt] = [i]
                }
            }

            for (const elt in populardict) {
                if (populardict.hasOwnProperty(elt)) {
                    delete b2j[elt]
                }
            }

            const isjunk = this.isjunk
            const junkdict = {}
            if (isjunk) {
                for (const elt in populardict) {
                    if (populardict.hasOwnProperty(elt) && isjunk(elt)) {
                        junkdict[elt] = 1
                        delete populardict[elt]
                    }
                }
                for (const elt in b2j) {
                    if (b2j.hasOwnProperty(elt) && isjunk(elt)) {
                        junkdict[elt] = 1
                        delete b2j[elt]
                    }
                }
            }

            this.isbjunk = difflib.isindict(junkdict)
            this.isbpopular = difflib.isindict(populardict)
        }

        this.find_longest_match = (alo, ahi, blo, bhi) => {
            const a = this.a
            const b = this.b
            const b2j = this.b2j
            const isbjunk = this.isbjunk
            let besti = alo
            let bestj = blo
            let bestsize = 0
            let j = null

            let j2len = {}
            const nothing = []
            for (let i = alo; i < ahi; i++) {
                const newj2len = {}
                const jdict = difflib.dictget(b2j, a[i], nothing)
                for (const jkey in jdict) {
                    if (jdict.hasOwnProperty(jkey)) {
                        let k
                        j = jdict[jkey]
                        if (j < blo) {
                            continue
                        }
                        if (j >= bhi) {
                            break
                        }
                        newj2len[j] = k = difflib.dictget(j2len, j - 1, 0) + 1
                        if (k > bestsize) {
                            besti = i - k + 1
                            bestj = j - k + 1
                            bestsize = k
                        }
                    }
                }
                j2len = newj2len
            }

            while (besti > alo && bestj > blo
                && !isbjunk(b[bestj - 1])
                && a[besti - 1] === b[bestj - 1]
            ) {
                besti--
                bestj--
                bestsize++
            }

            while (besti + bestsize < ahi && bestj + bestsize < bhi
                && !isbjunk(b[bestj + bestsize])
                && a[besti + bestsize] === b[bestj + bestsize]
            ) {
                bestsize++
            }

            while (besti > alo && bestj > blo
                && isbjunk(b[bestj - 1])
                && a[besti - 1] === b[bestj - 1]
            ) {
                besti--
                bestj--
                bestsize++
            }

            while (besti + bestsize < ahi && bestj + bestsize < bhi
                && isbjunk(b[bestj + bestsize])
                && a[besti + bestsize] === b[bestj + bestsize]
            ) {
                bestsize++
            }

            return [besti, bestj, bestsize]
        }

        this.get_matching_blocks = () => {
            if (this.matching_blocks != null) {
                return this.matching_blocks
            }
            const la = this.a.length
            const lb = this.b.length

            const queue = [
                [0, la, 0, lb]
            ]

            const matchingBlocks = []
            let alo
            let ahi
            let blo
            let bhi
            let qi
            let i
            let j
            let k
            let x
            while (queue.length) {
                qi = queue.pop()
                alo = qi[0]
                ahi = qi[1]
                blo = qi[2]
                bhi = qi[3]
                x = this.find_longest_match(alo, ahi, blo, bhi)
                i = x[0]
                j = x[1]
                k = x[2]

                if (k) {
                    matchingBlocks.push(x)
                    if (alo < i && blo < j) {
                        queue.push([alo, i, blo, j])
                    }
                    if (i + k < ahi && j + k < bhi) {
                        queue.push([i + k, ahi, j + k, bhi])
                    }
                }
            }

            matchingBlocks.sort(difflib.ntuplecomp)

            let i1 = 0
            let j1 = 0
            let k1 = 0
            let block = 0
            const nonAdjacent = []
            for (const idx in matchingBlocks) {
                if (matchingBlocks.hasOwnProperty(idx)) {
                    block = matchingBlocks[idx]
                    const i2 = block[0]
                    const j2 = block[1]
                    const k2 = block[2]
                    if (i1 + k1 === i2 && j1 + k1 === j2) {
                        k1 += k2
                    } else {
                        if (k1) {
                            nonAdjacent.push([i1, j1, k1])
                        }
                        i1 = i2
                        j1 = j2
                        k1 = k2
                    }
                }
            }

            if (k1) {
                nonAdjacent.push([i1, j1, k1])
            }

            nonAdjacent.push([la, lb, 0])
            this.matchingBlocks = nonAdjacent
            return this.matchingBlocks
        }

        this.getOpcodes = () => {
            if (this.opcodes != null) {
                return this.opcodes
            }
            let i = 0
            let j = 0
            const answer = []
            this.opcodes = answer
            let block
            let ai
            let bj
            let size
            let tag
            const blocks = this.get_matching_blocks()
            for (const idx in blocks) {
                if (blocks.hasOwnProperty(idx)) {
                    block = blocks[idx]
                    ai = block[0]
                    bj = block[1]
                    size = block[2]
                    tag = ''
                    if (i < ai && j < bj) {
                        tag = 'replace'
                    } else if (i < ai) {
                        tag = 'delete'
                    } else if (j < bj) {
                        tag = 'insert'
                    }
                    if (tag) {
                        answer.push([tag, i, ai, j, bj])
                    }

                    i = ai + size
                    j = bj + size

                    if (size) {
                        answer.push(['equal', ai, i, bj, j])
                    }
                }
            }

            return answer
        }

        // this is a generator function in the python lib, which of course is not supported in javascript
        // the reimplementation builds up the grouped opcodes into a list in their entirety and returns that.
        this.get_grouped_opcodes = n => {
            if (!n) {
                n = 3
            }
            let codes = this.getOpcodes()
            if (!codes) {
                codes = [
                    ['equal', 0, 1, 0, 1]
                ]
            }
            let code
            let tag
            let i1
            let i2
            let j1
            let j2
            if (codes[0][0] === 'equal') {
                code = codes[0]
                tag = code[0]
                i1 = code[1]
                i2 = code[2]
                j1 = code[3]
                j2 = code[4]
                codes[0] = [tag, Math.max(i1, i2 - n), i2, Math.max(j1, j2 - n), j2]
            }
            if (codes[codes.length - 1][0] === 'equal') {
                code = codes[codes.length - 1]
                tag = code[0]
                i1 = code[1]
                i2 = code[2]
                j1 = code[3]
                j2 = code[4]
                codes[codes.length - 1] = [tag, i1, Math.min(i2, i1 + n), j1, Math.min(j2, j1 + n)]
            }

            const nn = n + n
            let group = []
            const groups = []
            for (const idx in codes) {
                if (codes.hasOwnProperty(idx)) {
                    code = codes[idx]
                    tag = code[0]
                    i1 = code[1]
                    i2 = code[2]
                    j1 = code[3]
                    j2 = code[4]
                    if (tag === 'equal' && i2 - i1 > nn) {
                        group.push([tag, i1, Math.min(i2, i1 + n), j1, Math.min(j2, j1 + n)])
                        groups.push(group)
                        group = []
                        i1 = Math.max(i1, i2 - n)
                        j1 = Math.max(j1, j2 - n)
                    }

                    group.push([tag, i1, i2, j1, j2])
                }
            }

            if (group && !(group.length === 1 && group[0][0] === 'equal')) {
                groups.push(group)
            }

            return groups
        }

        this.ratio = () => {
            const matches = difflib.reduce((sum, triple) => {
                return sum + triple[triple.length - 1]
            }, this.get_matching_blocks(), 0)
            return difflib.calculate_ratio(matches, this.a.length + this.b.length)
        }

        this.quick_ratio = () => {
            let fullbcount
            let elt
            if (this.fullbcount === null) {
                this.fullbcount = fullbcount = {}
                for (let i = 0; i < this.b.length; i++) {
                    elt = this.b[i]
                    fullbcount[elt] = difflib.dictget(fullbcount, elt, 0) + 1
                }
            }
            fullbcount = this.fullbcount

            const avail = {}
            const availhas = difflib.isindict(avail)
            let matches = 0
            let numb = 0
            for (let i = 0; i < this.a.length; i++) {
                elt = this.a[i]
                if (availhas(elt)) {
                    numb = avail[elt]
                } else {
                    numb = difflib.dictget(fullbcount, elt, 0)
                }
                avail[elt] = numb - 1
                if (numb > 0) {
                    matches++
                }
            }

            return difflib.calculate_ratio(matches, this.a.length + this.b.length)
        }

        this.real_quick_ratio = () => {
            const la = this.a.length
            const lb = this.b.length
            return difflib.calculate_ratio(Math.min(la, lb), la + lb)
        }

        this.isjunk = isjunk || difflib.defaultJunkFunction
        this.a = this.b = null
        this.set_seqs(a, b)
    }
}

export default difflib
