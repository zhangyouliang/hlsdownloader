## # m3u8

    # 解密 media_0.ts => media_decryptd_0.ts
    strkey=$(hexdump -v -e '16/1 "%02x"' 1.key)
    iv=$(printf '%032x' 1)
    openssl aes-128-cbc -d -in media_0.ts -out media_decryptd_0.ts -nosalt -iv $iv -K $strkey
    
    
    # aes-128-cbc 
    # 加密解密过程测试
    openssl aes-128-cbc -d -in hello_en.txt -out hello_de.txt  -nosalt -iv $iv -K $strkey
    openssl aes-128-cbc -d -in hello_en.txt -out hello_de.txt  -nosalt -iv $iv -K $strkey