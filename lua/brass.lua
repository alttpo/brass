
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
        local tok = s:sub(ms, me-1)
        if tok:sub(1,1) == '@' then
            -- escaped token:
            return tok:sub(2), me, nil
        elseif tok == 'nil' then
            return nil, me, nil
        elseif tok == 'true' then
            return true, me, nil
        elseif tok == 'false' then
            return false, me, nil
        else
            -- regular token:
            return tok, me, nil
        end
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
