
local decode_list

local function decode_token(s, ms)
    local me = s:match('^[@%u%l_%./%?!][%u%l%d_%./%?!]*()', ms)
    if me == nil then
        return nil, ms, { err = 'unrecognized brass s-expression' }
    end
    return s:sub(ms, me-1), me, nil
end

local function decode_atom(s, ms)
    local m

    -- check if start of list:
    m = s:match('^()[%(]', ms)
    if m ~= nil then
        return decode_list(s, m)
    end

    -- check for token start:
    m = s:match('^()[@%u%l_%./%?!]', ms)
    if m ~= nil then
        return decode_token(s, m)
    end

    return nil, ms, { err = 'unrecognized brass s-expression' }
end

local function decode_list(s, ms)
    -- find balanced ( and ) positions:
    local ls, le = s:match('^()%b()()', ms)
    if le == nil then
        return nil, ms, { err = 'could not find end of list' }
    end

    local l = {}
    ms = ls+1
    while ms <= le do
        -- skip whitespace
        local we = s:match('^[% %\t]*()', ms)
        if we ~= nil then
            ms = we
        end

        -- end of list?
        if s:match('^[%)]', ms) == ')' then
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
