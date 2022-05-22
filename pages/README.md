# pages

pages enclose a viewport, a filter, and page data of a bespoke type

there is a decent amount of duplicated logic amidst pages, but i haven't yet been able
to come up with a clean generic interface for pages that dedupes this logic without 
running into too much trickyness

Could possibly have a "GetSelected" method exposed on the page based off the page viewport's Cursor location and filteredData
