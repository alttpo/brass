function fromhex_a1(s, len)
    local me = 1
    -- build string of bytes:
    local l = {}
    for i = 0,len-1 do
        local ok, x = pcall(tonumber, s:sub(me+i*2,me+i*2+1), 16)
        if not ok then
            return nil, me+i*2, { err = x }
        end
        l[#l+1] = string.char(x)
    end
    return table.concat(l), me, nil
end

function fromhex_a2(s, len)
    local me = 1
    -- build string of bytes:
    local l = {}
    for i = 0,len-1 do
        local ok, x = pcall(tonumber, s:sub(me+i*2,me+i*2+1), 16)
        if not ok then
            return nil, me+i*2, { err = x }
        end
        l[#l+1] = x
    end
    return string.char(unpack(l)), me, nil
end

function fromhex_b(s, len)
    local me = 1
    -- build string of bytes:
    local l = {}
    for i = 0,len-1 do
        local x = tonumber(s:sub(me+i*2,me+i*2+1), 16)
        l[#l+1] = x
    end
    return string.char(unpack(l)), me, nil
end

function fromhex_c(s, len)
    local me = 1
    -- build string of bytes:
    local l = {}
    for i = 0,len-1 do
        local x = tonumber(s:sub(me,me+1), 16)
        l[#l+1] = x
        me = me + 2
    end
    return string.char(unpack(l)), me, nil
end

function fromhex_d(s, len)
    local me = 1
    -- build string of bytes:
    local l = {}
    for i = 0,len-1 do
        l[#l+1] = tonumber(s:sub(me,me+1), 16)
        me = me + 2
    end
    return string.char(unpack(l)), me, nil
end

function fromhex_e(s, len)
    local me = 1
    -- build string of bytes:
    local l = {}
    for i = 0,len-1 do
        l[#l+1] = string.char(tonumber(s:sub(me,me+1), 16))
        me = me + 2
    end
    return table.concat(l), me, nil
end
