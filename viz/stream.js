/* global d3,rawdata */

const width = 960
const height = 500
const margin = ({ top: 0, right: 20, bottom: 30, left: 20 })
// .order(d3.stackOrderNone) // stackOrderNone stackOrderAscending  stackOrderDescending  stackOrderInsideOut
// .offset(d3.stackOffsetNone) // stackOffsetWiggle stackOffsetNone stackOffsetSilhouette

const svg = d3.select('body').append('svg')
  .attr('width', width)
  .attr('height', height)

// scales - x (no domain)
const x = d3.scaleTime()
  .range([margin.left, width - margin.right])

const y = d3.scaleLinear()
  .range([height - margin.bottom, margin.top])

// scales - color (no domain)
const color = d3.scaleOrdinal(d3.schemeCategory10)
// const color = d3.interpolateWarm // interpolateWarm interpolateCool

const xAxis = g => g
  .attr('transform', `translate(0,${height - margin.bottom})`)
  .call(d3.axisBottom(x).ticks(width / 80).tickSizeOuter(0))
  .call(g => g.select('.domain').remove())

const area = d3.area()
  .curve(d3.curveCardinal) // curveStep curveCardinal curveBasis
  .x(d => x(d.date))
  .y0(d => y(d.values[0]))
  .y1(d => y(d.values[1]))

// Below depend on data...
function render (data) {
  // clear the axix and area
  svg.selectAll('g').remove()

  x.domain(d3.extent(data, d => d.date))
  y.domain([d3.min(data, d => d.values[0]), d3.max(data, d => d.values[1])])
  color.domain(data.map(d => d.name))

  svg.append('g')
    .selectAll('path')
    .data([...multimap(data.map(d => [d.name, d]))])
    .enter().append('path')
    .attr('fill', ([name]) => color(name))
    .attr('d', ([, values]) => area(values))
    .append('title')
    .text(([name]) => name)

  // called once or every render ?
  svg.append('g')
    .call(xAxis)
}

fetchTransformAndDraw()
setInterval(async () => {
  console.log('Render!')
  fetchTransformAndDraw()
}, 5000)
// const data = mikedata() // sync version

async function fetchTransformAndDraw () {
  // const data = (await d3.json('https://raw.githubusercontent.com/vega/vega-lite/b2338345973f4717979ad9140c06ee0970c20116/data/unemployment-across-industries.json')).map(({ series, count, date }) => ({ name: series, value: count, date: new Date(date) }))
  const data = (await d3.json('data/unemployment-across-industries.json'))
    .map(({ series, count, date }) => ({ name: series, value: count, date: new Date(date) }))

  render(transform(data))
}

function transform (data) {
  // let data = rawdata.map(({ series, count, date }) => ({ name: series, value: count, date: new Date(date) }))

  // Compute the top nine industries, plus an “Other” category.
  const top = [...multisum(data.map(d => [d.name, d.value]))]
    .sort((a, b) => b[1] - a[1])
    .slice(0, 9)
    .map(d => d[0])
    .concat('Other')

  // Group the data by industry, then re-order the data by descending value.
  const series = multimap(data.map(d => [d.name, d]))
  data = [].concat(...top.map(name => series.get(name)))

  // Fold any removed (non-top) industries into the Other category.
  const other = series.get('Other')
  for (const [name, data] of series) {
    if (!top.includes(name)) {
      for (let i = 0, n = data.length; i < n; ++i) {
        if (+other[i].date !== +data[i].date) throw new Error()
        other[i].value += data[i].value
      }
    }
  }

  // Compute the stack offsets.
  const stack = d3.stack()
    .keys(top)
    .value((d, key) => d.get(key).value)
    // TODO order/offset from previous example
    // recommended insideOut, and wiggle
    .order(d3.stackOrderInsideOut) // stackOrderNone stackOrderAscending  stackOrderDescending  stackOrderInsideOut
    .offset(d3.stackOffsetWiggle)( // stackOffsetWiggle stackOffsetNone stackOffsetSilhouette
      Array.from(
        multimap(
          data.map(d => [+d.date, d]),
          (p, v) => p.set(v.name, v),
          () => new Map()
        ).values()
      ))

  // Copy the offsets back into the data.
  for (const layer of stack) {
    for (const d of layer) {
      d.data.get(layer.key).values = [d[0], d[1]]
    }
  }

  return data
}

function multimap (entries, reducer = (p, v) => (p.push(v), p), initializer = () => []) {
  const map = new Map()
  for (const [key, value] of entries) {
    map.set(key, reducer(map.has(key) ? map.get(key) : initializer(key), value))
  }
  return map
}

function multisum (entries) {
  return multimap(entries, (p, v) => p + v, () => 0)
}

// .attr('transform', 'translate(' + width / 2 + ',' + (height / 2) + ')')
// const svg = d3.select(DOM.svg(width, height))

function load () {
  // const gen = Array.from({ length: n }, () => bumps(m, k))
  const data = getdata()
  console.log('data', JSON.stringify(data))

  const xdomain = d3.extent(data, function (d) { return d.stamp })
  console.log('xdomain', xdomain)
  x.domain(xdomain)

  // console.log('const data = ', JSON.stringify(data))
  const nest = d3.nest().key(d => d.key).sortKeys(d3.ascending)

  const nesteddata = nest.entries(data)
  console.log('nested', nesteddata)
  console.log('nested', JSON.stringify(nesteddata))
  const layers = stack(nesteddata)
  console.log('layers', layers)

  y.domain([
    d3.min(layers, l => d3.min(l, d => d[0])),
    d3.max(layers, l => d3.max(l, d => d[1]))
  ])
  return layers
}

function getdata () {
// return [
//   [0.0002, 0.0027, 0.0253, 0.1528, 0.5999, 1.5311, 2.5392, 2.7374, 2.045, 2.2992, 1.8055, 0.2139, 0.0079, 0.0005, 0, 0, 0, 0, 0, 0],
//   [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.0005, 0.0268, 0.3715, 3.7822, 0.9674, 0.3052, 0.4762, 0.8833, 0.8757, 0.4566],
//   [0.046, 0.4105, 2.3163, 4.5502, 2.4683, 1.3599, 2.1985, 2.5416, 1.8129, 0.6976, 0.1447, 0.0162, 0.001, 0, 0, 0, 0, 0, 0.0002, 0.0172]
// ]
  if (0) {
    return [
      d3.range(18).map(x => Math.random()),
      d3.range(18).map(x => Math.random()),
      d3.range(18).map(x => Math.random()),
      [0, 1, 2, 1, 0, -1, -2, -1, 0, 1, 2, 1, 0, -1, -2, -1, 0, 1].map(x => (x + 2) / 4.1 + 0.1),
      [0, 1, 2, 1, 0, -1, -2, -1, 0, 1, 2, 1, 0, -1, -2, -1, 0, 1].map(x => (x + 2) / 4.1 + 0.1),
      [-1, 0, 1, 2, 1, 0, -1, -2, -1, 0, 1, 2, 1, 0, -1, -2, -1, 0].map(x => (-x + 2) / 4.1 + 0.1),
      [-2, -1, 0, 1, 2, 1, 0, -1, -2, -1, 0, 1, 2, 1, 0, -1, -2, -1].map(x => -x - 1)
      // [0.046, 0.4105, 2.3163, 4.5502, 2.4683, 1.3599, 2.1985, 2.5416, 1.8129, 0.6976, 0.1447, 0.0162, 0.001, 0, 0, 0, 0, 0, 0.0002, 0.0172]
    ]
  }

  return [
    // [
    { 'type': '', 'stamp': '2018-09-30', 'size': 34749262726, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-01', 'size': 104331733557, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-02', 'size': 117284001241, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-03', 'size': 107162458407, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-04', 'size': 123849484129, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-05', 'size': 85752783104, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-06', 'size': 134287301707, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-07', 'size': 70324783777, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-08', 'size': 679383597, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-09', 'size': 331409879, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-10', 'size': 613931767, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-11', 'size': 281689658, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-12', 'size': 566649309, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-13', 'size': 288803310, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-14', 'size': 399931108, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-15', 'size': 543700963, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-16', 'size': 284346440, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-17', 'size': 457778470, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-18', 'size': 671480830, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-19', 'size': 334514297, 'chunk': 0, 'fname': '/' },
    { 'type': '', 'stamp': '2018-10-20', 'size': 97378308, 'chunk': 0, 'fname': '/' },
    // ], [
    { 'type': '', 'stamp': '2018-09-30', 'size': 13205517070, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-01', 'size': 83984755223, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-02', 'size': 102178881489, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-03', 'size': 95624021609, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-04', 'size': 110950862270, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-05', 'size': 79400234070, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-06', 'size': 133921929618, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-07', 'size': 67151982904, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-08', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-09', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-10', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-11', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-12', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-13', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-14', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-15', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-16', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-17', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-18', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-19', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    { 'type': '', 'stamp': '2018-10-20', 'size': 1, 'chunk': 0, 'fname': '/Volumes/Space/' },
    // ], [
    { 'type': '', 'stamp': '2018-09-30', 'size': 21536239858, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-01', 'size': 20342983059, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-02', 'size': 15042581474, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-03', 'size': 11538436798, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-04', 'size': 12898621859, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-05', 'size': 6352549034, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-06', 'size': 365363496, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-07', 'size': 3133198621, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-08', 'size': 679383596, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-09', 'size': 331379203, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-10', 'size': 613931766, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-11', 'size': 281683572, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-12', 'size': 566649308, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-13', 'size': 288803309, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-14', 'size': 399931107, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-15', 'size': 543680297, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-16', 'size': 284262333, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-17', 'size': 457770976, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-18', 'size': 671474071, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-19', 'size': 334514296, 'chunk': 0, 'fname': '/Users/daniel/' },
    { 'type': '', 'stamp': '2018-10-20', 'size': 97378308, 'chunk': 0, 'fname': '/Users/daniel/' },
    // ], [
    { 'type': '', 'stamp': '2018-09-30', 'size': 32082, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-01', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-02', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-03', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-04', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-05', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-06', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-07', 'size': 1701242606, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-08', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-09', 'size': 18611432, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-04', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-11', 'size': 19325506, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-10', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-13', 'size': 37330926, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-12', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-15', 'size': 106236792, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-16', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-17', 'size': 82968844, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-18', 'size': 40548, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-19', 'size': 96675164, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-20', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' }
    // ]
  ].map(d => ({ key: d.fname, value: d.size, date: new Date(d.stamp) }))
}
