
local decode_list

local function decode_atom(s, ms)
    local m, me

    -- check if start of list:
    m = s:match('^()[%(]', ms)
    if m ~= nil then
        return decode_list(s, m)
    end

    -- check for token:
    me = s:match('^[@%u%l_%./%?!][%u%l%d_%./%?!]*()', ms)
    if me ~= nil then
        local v = s:sub(ms, me-1)
        if v:sub(1,1) == '@' then
            -- escaped token:
            return v:sub(2), me, nil
        elseif v == 'nil' then
            return nil, me, nil
        elseif v == 'true' then
            return true, me, nil
        elseif v == 'false' then
            return false, me, nil
        else
            -- regular token:
            return v, me, nil
        end
    end

    -- check for decimal integer:
    me = s:match('^%-?%d+()', ms)
    if me ~= nil then
        local v = s:sub(ms, me-1)
        return tonumber(v, 10), me, nil
    end

    -- check for hexadecimal integer:
    me = s:match('^%-?%$[0-9a-f]+()', ms)
    if me ~= nil then
        local v = s:sub(ms, me-1)
        if v:sub(1,1) == '-' then
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

        -- build string of bytes:
        local l = {}
        for i = 0,len-1 do
            l[#l+1] = tonumber(s:sub(me,me+1), 16)
            me = me + 2
        end

        return string.char(unpack(l)), me, nil
    end

    return nil, ms, { err = 'unrecognized brass s-expression' }
end

decode_list = function (s, ms)
    -- find balanced ( and ) positions:
    local le = s:match('^%b()()', ms)
    if le == nil then
        return nil, ms, { err = 'could not find end of list' }
    end

    local l = {}
    ms = ms + 1
    while ms <= le do
        -- skip whitespace
        local we = s:match('^[% %\t]*()', ms)
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

function brass_decode(s)
    local expr, me, err = decode_list(s, 1)
    return expr, me, err
end
