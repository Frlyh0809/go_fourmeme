
## manager1
>0xEC4549caDcE5DA21Df6E6422d448034B5233bFbC

manager1使用已经很少了，10天内交易不到20次，解析下面两种足够

```aiignore

https://bscscan.com/tx/0x4237f79527c3021392a91d4c5040def13df82910a89ad2bc9baadc8fc53dcf86#eventlog
    指令：saleToken 0x9b911b5e
    推测行为:sell
    只有一个Transfer
    log[0] Transfer:0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
        address=token
        to=manager1
        value=tokenTransfer_amount
    log[2] 0x80d4e495cda89b31af98c8e977ff11f417bafcee26902a17a15be51830c47533
        address=manager1
        value[0]=token
        value[1]=buyer
        value[2]=tokenTransfer_amount
        value[3]=bnbTransfer_amount
        value[4]=bnbTransfer_fee_amount
```

```aiignore
https://bscscan.com/tx/0xa31f2342ea7e3c28374835f45ac388e0f5707ade75bd9c8c406bef4a46845fa5#eventlog
    指令：0x3deec419
    推测行为:buy
    只有一个Transfer
    log[0] Transfer:0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
        address=token
        from=manager1
        value=tokenTransfer_amount
    log[2] TokenPurchase:0x00fe0e12b43090c1fc19a34aefa5cc138a4eeafc60ab800f855c730b3fb9480e
        address=manager1
        value[0]=token
        value[1]=buyer
        value[2]=tokenTransfer_amount
        value[3]=bnbTransfer_amount
        value[4]=bnbTransfer_fee_amount
```
## TokenManager2
>0x5c952063c7fc8610FFDB798152D69F0B9550762b
>>DefaultFourmemeManager

```aiignore
https://bscscan.com/tx/0x5398b39bb98be01c8ec4259644cfc93d8213a207fb6c7ab763502fdea198a9ee
Function: createToken(bytes code, bytes poolsCode) ***
0x519ebb10

0x396d5e902b675b032348d3d2e9517ee8f0c4a926603fbc075d3d282ff00cad20
```


```aiignore
https://bscscan.com/tx/0x98f1ee9749d304217a4c2fee48994dae856d60dd7551494edad8702f3f652f63
Function: buyMemeToken(address tokenManager,address token,address recipient,uint256 funds,uint256 minAmount)
0x7771fdb0

0x7db52723a3b2cdd6164364b3b766e65e540d7be48ffa89582956d8eaebe62942
```
```aiignore
https://bscscan.com/tx/0x8d637ccfc01ada23129b24452c74cdfb2f2c057bc83fe1d73d16eff1cd9ab43f
Function: smartSwapTo(uint256 orderId,address receiver,tuple baseRequest,uint256[] batchesAmount,tuple[][] batches,tuple[] )
0x03b87e5f

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19
```

```aiignore
https://bscscan.com/tx/0x13c5d421bb590d4e3fe635a216f7639528ea18430b561803cc6ee8f69a6314a4
Function: swapSell(address _token,uint256 _percent,uint256 minamount,uint256 tokenType,address[] v2path,bytes v3path) ***
0x389afe6f

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19
```

```aiignore
https://bscscan.com/tx/0x660559939d9ea286d9b6a37766bfa62ef3ffdfb581597c209f47255e042a07f3
未知
0xe63aaf36

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19
```

```aiignore
https://bscscan.com/tx/0xd02a910ba01c1cb2ab6fce11b14cb9703d49ff27ecab057974dd4c3a574251ed
未知
0x06e7b98f


```

```aiignore
https://bscscan.com/tx/0xf862b3e3e599f15e842140c8828c48fbc3e796f28075ff25af2687e7aa26f297
Function: axiomTrade(bytes commands,bytes[] inputs,uint256 deadline) ***
0x3e0f9c3c

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19

```
```aiignore
https://bscscan.com/tx/0xb4f60b0b2ac66019ea731bd61e42a2acecd6c5a9343421732fe8d495750c7511
0x47ee97ff

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19
```

```aiignore
https://bscscan.com/tx/0x3f78d58536bcd259242ae9b4c98cd37864506f1286825437e3f1f503f8c3a864
Function: sellToken(address userAddress, uint256 tokenQty) ***
0xf464e7db

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19
```

```aiignore
https://bscscan.com/tx/0x752ec9a78a8f6bdae5ec97525dca350cd916413aad1b2d142756a565659cec1e
未知
0xedf9e251

0x7db52723a3b2cdd6164364b3b766e65e540d7be48ffa89582956d8eaebe62942
```

```aiignore
https://bscscan.com/tx/0xb690fb77c1b3c27180fb4b51d7e2f4003b10847ac76cc6cd3fca746399d87b54
Function: buyTokenAMAP(address token,uint256 funds,uint256 minAmount) ***
0x87f27655

0x7db52723a3b2cdd6164364b3b766e65e540d7be48ffa89582956d8eaebe62942

```

```aiignore
https://bscscan.com/tx/0xa99845667d51d9a353f3813b997ec807fdfbc08d300309954c189b6c8ce41397
Function: sellToken(uint256 origin,address token,uint256 amount,uint256 minFunds) ***
0x0da74935

0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19

```


## TokenManagerHelper3
>0xF251F83e40a78868FcfA3FA4599Dad6494E46034