package gjson

import (
	"strconv"
)

const unescapeStackBufSize = 64

type callback func(node *PathNode, value Value)
type parseManyContext struct {
	json []byte

	cb callback

	path *PathNode

	checked int
}

func ParseMany(json []byte, path *PathNode, cb callback) {
	var i int
	var c = &parseManyContext{
		json: json, path: path, cb: cb,
	}
	for ; i < len(c.json); i++ {
		if json[i] == '{' {
			i++
			parseObjectMany(c, i, path)
			break
		}
		if json[i] == '[' {
			i++
			parseArrayMany(c, i, path)
			break
		}
	}
}

func parseObjectMany(c *parseManyContext, i int, path *PathNode) (int, bool) {
	var kesc, vesc, ok bool
	var key []byte
	var val []byte

	var stackBuf [unescapeStackBufSize]byte

	for i < len(c.json) {
		// parse key
		for ; i < len(c.json); i++ {
			if c.json[i] == '"' {
				i++
				var s = i
				for ; i < len(c.json); i++ {
					if c.json[i] > '\\' {
						continue
					}
					if c.json[i] == '"' {
						i, key, kesc, ok = i+1, c.json[s:i], false, true
						goto parseKeyStringDone
					}
					if c.json[i] == '\\' {
						i++
						for ; i < len(c.json); i++ {
							if c.json[i] > '\\' {
								continue
							}
							if c.json[i] == '"' {
								// look for an escaped slash
								if c.json[i-1] == '\\' {
									n := 0
									for j := i - 2; j > 0; j-- {
										if c.json[j] != '\\' {
											break
										}
										n++
									}
									if n%2 == 0 {
										continue
									}
								}
								i, key, kesc, ok = i+1, c.json[s:i], true, true
								goto parseKeyStringDone
							}
						}
						break
					}
				}
				key, kesc, ok = c.json[s:], false, false
			parseKeyStringDone:
				break
			}

			// end of object
			if c.json[i] == '}' {
				return i + 1, true
			}
		}
		if !ok {
			return i, false
		}
		if kesc {
			key, _ = unescapeBytes(key, stackBuf[:])
		}

		next, ok := path.child[string(key)]
		hit, subChk := ok && next.check, ok && next.sub > 0

		// parse value
		for ; i < len(c.json); i++ {
			switch c.json[i] {
			default:
				continue
			case '"':
				i++
				if i, val, vesc, ok = parseStringBytes(c.json, i); !ok {
					return i, false
				}

				if hit {

					var value = Value{
						Raw:  val,
						Type: String,
						Str:  string(val[1 : len(val)-1]),
					}
					if vesc {
						value.Str = unescape(value.Str)
					}

					_ = value
					c.cb(next, value)
					c.checked += next.sub + 1
				}
			case '{':
				if subChk {
					i, _ = parseObjectMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i)
				}

				if hit {
					c.cb(next, Value{Raw: val, Type: Object})
					c.checked++
				}
			case '[':
				if hit {
					i, _ = parseArrayMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i)
				}

				if hit {
					c.cb(next, Value{Raw: val, Type: Array})
					c.checked++
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				var typ Type
				i, val, typ = parseNumberByte(c.json, i)
				if hit {
					c.cb(next, Value{Raw: val, Type: typ})
					c.checked += next.sub + 1
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteralByte(c.json, i)
				if hit {
					var value = Value{Raw: val}
					switch vc {
					case 't':
						value.Type = True
					case 'f':
						value.Type = False
					case 'n':
						value.Type = Null
					}
					c.cb(next, value)
					c.checked += next.sub + 1
				}
			}
			break
		}

		if c.checked >= c.path.sub {
			return i, true
		}
	}

	return i, true
}

func parseArrayMany(c *parseManyContext, i int, path *PathNode) (int, bool) {
	var vesc, hit bool
	var val []byte
	var idx int

	for ; i < len(c.json)+1; idx++ {
		next, ok := path.child[strconv.Itoa(idx)]
		hit = ok && next.check

		for ; ; i++ {
			var ch byte
			if i > len(c.json) {
				break
			} else if i == len(c.json) {
				ch = ']'
			} else {
				ch = c.json[i]
			}
			switch ch {
			default:
				continue
			case '"':
				i++
				if i, val, vesc, ok = parseStringBytes(c.json, i); !ok {
					return i, false
				}

				if hit {
					var value = Value{
						Raw:  val,
						Type: String,
						Str:  string(val[1 : len(val)-1]),
					}
					if vesc {
						value.Str = unescape(value.Str)
					}

					c.cb(next, value)
					c.checked += next.sub + 1
				}
			case '{':
				if hit && next.sub > 0 {
					i, _ = parseObjectMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i+1)
				}

				if hit {
					c.cb(next, Value{Raw: val, Type: Object})
					c.checked++
				}
			case '[':
				if hit {
					i, _ = parseArrayMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i)
				}

				if hit {
					c.cb(next, Value{Raw: val, Type: Array})
					c.checked++
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				var typ Type
				i, val, typ = parseNumberByte(c.json, i)
				if hit {
					c.cb(next, Value{Raw: val, Type: typ})
					c.checked += next.sub + 1
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteralByte(c.json, i)
				if hit {
					var value = Value{Raw: val}
					switch vc {
					case 't':
						value.Type = True
					case 'f':
						value.Type = False
					case 'n':
						value.Type = Null
					}
					c.cb(next, value)
					c.checked += next.sub + 1
				}
			case ']':
				return i + 1, true
			}
			break
		}
	}

	return i, true
}

func parseStringBytes(json []byte, i int) (int, []byte, bool, bool) {
	var s = i
	for ; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return i + 1, json[s-1 : i+1], false, true
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					// look for an escaped slash
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					return i + 1, json[s-1 : i+1], true, true
				}
			}
			break
		}
	}
	return i, json[s-1:], false, false
}

func parseNumberByte(json []byte, i int) (int, []byte, Type) {
	var s = i
	i++
	var typ = Integer
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' || json[i] == '}' {
			return i, json[s:i], typ
		} else if json[i] == '.' {
			typ = Float
		}
	}
	return i, json[s:], typ
}

func parseLiteralByte(json []byte, i int) (int, []byte) {
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

func parseSquashByte(json []byte, i int) (int, []byte) {
	s := i
	i++
	depth := 1
	for ; i < len(json); i++ {
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					if json[i] == '"' {
						// look for an escaped slash
						if json[i-1] == '\\' {
							n := 0
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							if n%2 == 0 {
								continue
							}
						}
						break
					}
				}
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					i++
					return i, json[s:i]
				}
			}
		}
	}
	return i, json[s:]
}
