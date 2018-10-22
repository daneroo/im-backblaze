/* global d3 */

const width = 960
const height = 500
const margin = ({ top: 0, right: 20, bottom: 30, left: 20 })

const svg = d3.select('body').append('svg')
  .attr('width', width)
  .attr('height', height)
  .style('font', '10px sans-serif')

// scales - x (no domain)
const x = d3.scaleTime()
  .range([margin.left, width - margin.right])

// scales - y (no domain)
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
  .curve(d3.curveBasis) // curveStep curveCardinal curveBasis
  .x(d => x(d.date))
  .y0(d => y(d.values[0]))
  .y1(d => y(d.values[1]))

// Below depend on data...
function render (data) {
  // clear the x-axis and area
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

  legend()
}

function legend () {
  const g = svg.append('g')
    .selectAll('g')
    .data(color.domain().slice()/* .reverse() */)
    .enter().append('g')
    .attr('transform', (d, i) => `translate(${margin.left},${i * 20})`)

  g.append('rect')
    .attr('width', 19)
    .attr('height', 19)
    .attr('fill', color)

  g.append('text')
    .attr('x', 24)
    .attr('y', 9.5)
    .attr('dy', '0.35em')
    .text(d => d)
}

fetchTransformAndDraw()
const intvl = setInterval(async () => {
  console.log('Render!')
  fetchTransformAndDraw()
  clearInterval(intvl)
}, 5000)

async function fetchTransformAndDraw () {
  // const data = await unemploymentData()
  const data = await bzData()

  // In the file, this is the structure:
  // const infile = [
  //   { 'series': 'Government', 'year': 2000, 'month': 1, 'count': 430, 'rate': 2.1, 'date': '2000-01-01T08:00:00.000Z' },
  //   { 'series': 'Government', 'year': 2000, 'month': 2, 'count': 409, 'rate': 2, 'date': '2000-02-01T08:00:00.000Z' }
  // ]
  // after load, this is the structure:
  // const after = [
  //   { 'name': 'Government', 'value': 430, 'date': '2000-01-01T08:00:00.000Z' },
  //   { 'name': 'Government', 'value': 409, 'date': '2000-02-01T08:00:00.000Z' }
  // ]

  render(transform(data))
}

function transform (data) {
  const N = 9
  // Compute the top N industries, plus an “Other” category.
  const top = [...multisum(data.map(d => [d.name, d.value]))]
    .sort((a, b) => b[1] - a[1])
    .slice(0, N)
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
    .order(d3.stackOrderAscending) // stackOrderNone stackOrderAscending  stackOrderDescending  stackOrderInsideOut
    .offset(d3.stackOffsetSilhouette)( // stackOffsetWiggle stackOffsetNone stackOffsetSilhouette
      Array.from(
        multimap(
          data.map(d => [+d.date, d]),
          (p, v) => p.set(v.name, v),
          () => new Map()
        ).values()
      ))

  console.log('stack', stack)

  // Copy the offsets back into the data.
  for (const layer of stack) {
    for (const d of layer) {
      d.data.get(layer.key).values = [d[0], d[1]]
    }
  }

  return data
}

function multimap (entries, reducer = (p, v) => ((p.push(v), p)), initializer = () => []) {
  const map = new Map()
  for (const [key, value] of entries) {
    map.set(key, reducer(map.has(key) ? map.get(key) : initializer(key), value))
  }
  return map
}

function multisum (entries) {
  return multimap(entries, (p, v) => p + v, () => 0)
}

async function unemploymentData () {
  // const dataURL = 'https://raw.githubusercontent.com/vega/vega-lite/b2338345973f4717979ad9140c06ee0970c20116/data/unemployment-across-industries.json'
  const dataURL = 'data/unemployment-across-industries.json'
  let data = (await d3.json(dataURL))
    .map(({ series, count, date }) => ({ name: series, value: count, date: new Date(date) }))

  return data
}

function parents (path, depth) {
  const parts = path.split('/')
  return parts.slice(0, depth + 1).join('/')
}

async function bzData () {
  const dataURL = 'data/allxfrs.json'
  let data = (await d3.json(dataURL))
    .map(({ fname, size, stamp }) => ({ name: fname, value: size, date: new Date(stamp) }))

  console.log(`Fetched ${data.length} entries`)
  const summary = {
    Other: { name: 'Other', value: 0, count: 0, byDate: {} }
  }
  const alldates = {}
  data.forEach((d, i) => {
    // console.log('i', i, d)
    const key = parents(d.name, 3)
    // 2018-10-21 12:37:01
    const dt = d.date.toISOString().substring(0, 10)
    // const dt = d.date.toISOString().substring(0, 13) + ':00:00'

    if (!(key in summary)) {
      summary[key] = { name: key, value: 0, count: 0, byDate: {} }
    }
    if (!(dt in summary[key].byDate)) {
      summary[key].byDate[dt] = { name: key, date: dt, value: 0 }
    }
    summary[key].count += 1
    summary[key].value += d.value
    summary[key].byDate[dt].value += d.value

    if (!(dt in alldates)) {
      alldates[dt] = { date: dt, count: 0 }
    }
    alldates[dt].count += 1
  })
  // console.log('summary', JSON.stringify(summary, null, 2))
  // console.log('alldates', JSON.stringify(alldates, null, 2))
  const ds = Object.keys(alldates).sort()
  // console.log('ds', JSON.stringify(ds, null, 2))

  const rdata = []
  for (let key in summary) {
    // console.log(`--${key}:${Object.keys(summary[key].byDate).sort()}`)
    for (let dt of ds) {
      if (!(dt in summary[key].byDate)) {
        // console.log(`Inserting date ${dt} in ${key}`)
        summary[key].byDate[dt] = { name: key, date: dt, value: 0 }
        rdata.push({ name: key, date: dt, value: 0 })
      } else {
        // console.log(`Found date ${dt} in ${key}: ${JSON.stringify(summary[key].byDate[dt])}`)
        rdata.push(summary[key].byDate[dt])
      }
    }
  }
  // rdata.forEach((d, i, rdata) => {
  //   console.log(`${i}/${rdata.length}: ${JSON.stringify(d)}`)
  // })
  if (rdata.length > 0) {
    return rdata
      .map(d => ({ ...d, date: new Date(d.date) }))
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
    { 'type': '', 'stamp': '2018-10-10', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-11', 'size': 19325506, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-12', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-13', 'size': 37330926, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-14', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-15', 'size': 106236792, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-16', 'size': 1, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-17', 'size': 82968844, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-18', 'size': 40548, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-19', 'size': 96675164, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' },
    { 'type': '', 'stamp': '2018-10-20', 'size': 42, 'chunk': 0, 'fname': '/Users/daniel/Library/Containers/com.docker.docker/Data/' }
    // ]
  ]
    // .filter(d => d.stamp > '2018-10-07')
    .map(d => ({ name: d.fname, value: d.size, date: new Date(d.stamp) }))
    // .map(d => ({ ...d, value: Math.log(d.value) }))
}
