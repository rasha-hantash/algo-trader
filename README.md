# algo-trader

We want to retroactively monitor our trades
- Real-time web socket market data 
- Engine that computes what trades should, be will send the trade 
- Execute 
    - Minimize slippage (this is a constant number) 
        - Order book
            - Bids (what are people will to buy it for )
            - Offer -> what are they wiling to sell it for) 
            - Should also be consuming order 
            - Real-time price = caulcation of bid and asks 
        - Slippage 
            - Average price gets driven up inadvenrtantly 
        - Front running orders 
        - Minimize slippage (my ask and what is the 2% of the ask 


MarketData Func() 
- Parse Symbol
- BidSorted[symbol]
- Update()
    - Upsert()
    - Delete() 


- Store the risk thresholds in a tree 
- Dollar CNH is at 7.2 
    - 5 diff order with risk threshold of 7.1, 7.1, 7.0
        - Store them as a tree 
        - Root of the tree (highest risk threshold) 