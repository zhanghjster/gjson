package gjson

import "strconv"

type parseManyContext struct {
	json []byte

	cb   func()

	path *PathNode

	checked int
}

func ParseMany(json []byte, path *PathNode, cb func()) {
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
	var kesc, vesc, ok, hit bool
	var key string
	var val []byte

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
						i, key, kesc, ok = i+1, string(c.json[s:i]), false, true
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
								i, key, kesc, ok = i+1, string(c.json[s:i]), true, true
								goto parseKeyStringDone
							}
						}
						break
					}
				}
				key, kesc, ok = string(c.json[s:]), false, false
			parseKeyStringDone:
				break
			}

			// end of object
			if c.json[i] == '}' {
				return i + 1, false
			}
		}
		if !ok {
			return i, false
		}
		if kesc {
			key = unescape(key)
		}

		next, ok := path.child[key]
		hit = ok && next.check

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
						Str: string(val[1 : len(val)-1]),
						Raw:val,
						Type: String,
					}
					if vesc {
						value.Str = unescape(value.Str)
					}

					c.cb()

					c.checked += next.sub + 1
				}
			case '{':
				if hit && next.sub > 0 {
					i, _ = parseObjectMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i+1)
				}

				if hit {
					c.checked++

					var value Value
					value.Raw = val
					value.Type = JSON

					c.cb()
				}
			case '[':
				if hit {
					i = parseArrayMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i)
				}

				if hit {
					c.checked++

					var value Value
					value.Raw = val
					value.Type = JSON

					c.cb()
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumberByte(c.json, i)
				if hit {
					var value Value

					value.Raw = val
					value.Type = Number

					c.checked += next.sub + 1

					c.cb()
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteralByte(c.json, i)
				if hit {
					var value Value
					value.Raw = val
					switch vc {
					case 't':
						value.Type = True
					case 'f':
						value.Type = False
					}

					c.checked += next.sub + 1

					c.cb()
				}
			}
			break
		}
	}

	return i, false
}

func parseArrayMany(c *parseManyContext, i int, path *PathNode) int {
	var vesc, hit bool
	var val []byte
	var h int

	for i < len(c.json)+1 {

		next, ok := path.child[strconv.Itoa(h)]
		hit = ok && next.check

		h++
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
				i, val, vesc, ok = parseStringBytes(c.json, i)
				if !ok {
					return i
				}
				if hit {
					var value Value
					if vesc {
						value.Str = unescape(string(val[1 : len(val)-1]))
					} else {
						value.Str = string(val[1 : len(val)-1])
					}
					value.Raw = val
					value.Type = String

					c.cb()

					c.checked++

					return i
				}
			case '{':
				if hit && next.sub > 0 {
					i, _ = parseObjectMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i+1)
				}

				if hit {
					c.checked++

					var value Value
					value.Raw = val
					value.Type = JSON

					c.cb()
				}
			case '[':
				if hit {
					i = parseArrayMany(c, i+1, next)
				} else {
					i, val = parseSquashByte(c.json, i)
				}

				if hit {
					c.checked++

					var value Value
					value.Raw = val
					value.Type = JSON

					c.cb()
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumberByte(c.json, i)
				if hit {
					var value Value

					value.Raw = val
					value.Type = Number

					c.checked += next.sub + 1

					c.cb()

					return i
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteralByte(c.json, i)
				if hit {
					var value Value
					value.Raw = val
					switch vc {
					case 't':
						value.Type = True
					case 'f':
						value.Type = False
					}

					c.checked += next.sub + 1

					c.cb()

					return i
				}
			case ']':
				return i + 1
			}
			break
		}
	}
	return i
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

func parseNumberByte(json []byte, i int) (int, []byte) {
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' || json[i] == '}' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
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
