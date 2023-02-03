-- Brass: a custom s-expression encoder and decoder library for Lua 5.1
-- Version 20230128
--
-- Copyright jsd1982 2023
--
-- MIT License
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
-- documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
-- rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
-- permit persons to whom the Software is furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
-- Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
-- WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
-- OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
-- OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

local brass = {}

-- lua 5.2 to 5.1 compat:
if table.unpack ~= nil then
    unpack = table.unpack
end

local decode_list

local function decode_atom(s, ms)
    local m, me

    -- check if start of list:
    m = s:match('^()[%(]', ms)
    if m ~= nil then
        return decode_list(s, m)
    end

    -- check for nil:
    me = s:match('^nil()', ms)
    if me ~= nil then
        return { __brass_kind = 'nil' }, me, nil
    end

    -- check for true:
    me = s:match('^true()', ms)
    if me ~= nil then
        return true, me, nil
    end

    -- check for false:
    me = s:match('^false()', ms)
    if me ~= nil then
        return false, me, nil
    end

    -- check for hexadecimal integer:
    me = s:match('^[%-]?%$[0-9a-f]+()', ms)
    if me ~= nil then
        local v = s:sub(ms, me-1)
        local g = v:sub(1,1)
        if g == '-' then
            return -tonumber(v:sub(3), 16), me, nil
        else
            return tonumber(v:sub(2), 16), me, nil
        end
    end

    -- check for hex-octets:
    me = s:match('^#[0-9a-f]+%$()', ms)
    if me ~= nil then
        -- parse length in hex:
        local len = tonumber(s:sub(ms+1, me-2), 16)

        -- extract hex digits:
        local he = me-1+len*2
        if he > #s then
            return nil, ms, { err = 'hex-octet sequence length incorrect' }
        end

        -- build list of octet values:
        local l = {}
        l.__brass_kind = 'octets'

        for i = 0,len-1 do
            l[#l+1] = tonumber(s:sub(me,me+1), 16)
            me = me + 2
        end

        return l, me, nil
    end

    -- check for string:
    me = s:match('^"[^"\\\r\n]*"()', ms)
    if me ~= nil then
        -- trivial string with no escaped chars:
        return s:sub(ms+1,me-2), me, nil
    elseif s:sub(ms, ms) == '"' then
        -- more complex string with escaped chars:
        ms = ms + 1
        local l = {}
        while ms <= #s do
            me = s:match('^[^"\\\r\n]+()', ms)
            if me ~= nil then
                l[#l+1] = s:sub(ms,me-1)
            else
                me = ms
            end

            local ec = s:sub(me,me)
            if ec == '"' then
                return table.concat(l), me+1, nil
            elseif ec == '\\' then
                -- handle escapes:
                ms = me + 1
                local hx = s:match('x([0-9a-f][0-9a-f])', ms)
                if hx ~= nil then
                    ms = ms + 3
                    l[#l+1] = string.char(tonumber(hx,16))
                else
                    ec = s:sub(ms,ms)
                    if ec == 't' then
                        l[#l+1] = '\t'
                    elseif ec == 'r' then
                        l[#l+1] = '\r'
                    elseif ec == 'n' then
                        l[#l+1] = '\n'
                    elseif ec == '\\' then
                        l[#l+1] = '\\'
                    elseif ec == '"' then
                        l[#l+1] = '"'
                    else
                        return nil, ms, { err = 'invalid escape sequence' }
                    end
                    ms = ms + 1
                end
            else
                return nil, me, { err = 'invalid string literal' }
            end
        end
    end

    return nil, ms, { err = 'unrecognized brass s-expression' }
end

decode_list = function (s, ms)
    local l = {}
    l.__brass_kind = 'list'

    ms = ms + 1
    while ms <= #s do
        -- skip whitespace
        local we = s:match('^[% ]*()', ms)
        if we ~= nil then
            ms = we
        end

        -- end of list?
        if s:sub(ms, ms) == ')' then
            return l, ms+1, nil
        end

        -- decode list item:
        local child, me, err = decode_atom(s, ms)
        if err ~= nil then
            return l, me, err
        end
        ms = me

        l[#l+1] = child
    end

    return l, me, { err = 'unexpected end of list' }
end

function brass.decode(s)
    local expr, me, err = decode_list(s, 1)
    return expr, me, err
end

function brass.encode(e)
    if e == nil then
        return 'nil'
    elseif e == true then
        return 'true'
    elseif e == false then
        return 'false'
    elseif type(e) == 'string' then
        local s = e
        -- escape characters:
        return '"' .. s:gsub('[^%w ]', function (m)
            local b = string.byte(m)
            if b == 9 then
                return '\\t'
            elseif b == 10 then
                return '\\n'
            elseif b == 13 then
                return '\\r'
            elseif b == 34 then
                return '\\"'
            elseif b == 92 then
                return '\\\\'
            elseif b < 32 or b >= 128 then
                return string.format('\\x%02x', b)
            else
                return m
            end
        end) .. '"'
    elseif type(e) == 'number' then
        if e < 0 then
            return string.format('-$%x', -e)
        else
            return string.format('$%x', e)
        end
    elseif type(e) == 'table' then
        if e.__brass_kind == 'nil' then
            return 'nil'
        elseif e.__brass_kind == 'list' then
            local l = {}
            for i=1,#e do
                l[#l+1] = brass.encode(e[i])
                l[#l+1] = ' '
            end
            if #l > 0 then
                l[#l] = nil
            end
            return '(' .. table.concat(l) .. ')'
        elseif e.__brass_kind == 'octets' then
            local l = {}
            for i=1,#e do
                l[#l+1] = string.format('%02x', e[i])
            end
            return '#' .. string.format('%x', #e) .. '$' .. table.concat(l)
        elseif e.__brass_kind == 'map' then
            local l = {}
            for k,v in pairs(e) do
                if k ~= '__brass_kind' then
                    l[#l+1] = '('
                    l[#l+1] = brass.encode(k)
                    l[#l+1] = ' '
                    l[#l+1] = brass.encode(v)
                    l[#l+1] = ')'
                end
            end
            return '{' .. table.concat(l, ' ') .. '}'
        end
    end
end

return brass
