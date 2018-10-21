
/* global d3 */

const width = 960
const height = 700
const radius = (Math.min(width, height) / 2) - 10
const formatNumber = d3.format(',d')

const x = d3.scaleLinear()
  .range([0, 2 * Math.PI])

const y = d3.scaleSqrt()
  .range([0, radius])

const color = d3.scaleOrdinal(d3.schemeCategory20)

const partition = d3.partition()

const arc = d3.arc()
  .startAngle(function (d) { return Math.max(0, Math.min(2 * Math.PI, x(d.x0))) })
  .endAngle(function (d) { return Math.max(0, Math.min(2 * Math.PI, x(d.x1))) })
  .innerRadius(function (d) { return Math.max(0, y(d.y0)) })
  .outerRadius(function (d) { return Math.max(0, y(d.y1)) })

const svg = d3.select('body').append('svg')
  .attr('width', width)
  .attr('height', height)
  .append('g')
  .attr('transform', 'translate(' + width / 2 + ',' + (height / 2) + ')')

function load (path) {
  d3.json(path, function (error, root) {
    if (error) throw error

    // clear first
    svg.selectAll('path').remove()

    root = d3.hierarchy(root)
    root.sum(function (d) { return d.size })
    console.log(root)
    svg.selectAll('path')
      .data(partition(root).descendants())
      .enter().append('path')
      .attr('d', arc)
      .style('fill', function (d) { return color((d.children ? d : d.parent).data.name) })
      .on('click', click)
      .append('title')
      .text(function (d) { return d.data.name + '\n' + formatNumber(d.value) })
  })
}

function click (d) {
  console.log(d.data.name)
  svg.transition()
    .duration(750)
    .tween('scale', function () {
      const xd = d3.interpolate(x.domain(), [d.x0, d.x1])
      const yd = d3.interpolate(y.domain(), [d.y0, 1])
      const yr = d3.interpolate(y.range(), [d.y0 ? 20 : 0, radius])
      return function (t) { x.domain(xd(t)); y.domain(yd(t)).range(yr(t)) }
    })
    .selectAll('path')
    .attrTween('d', function (d) { return function () { return arc(d) } })
}

// This is for the selection DropDown
const select = d3.select('body')
  .append('select')
  .attr('class', 'select')
  .on('change', onchange)

select
  .selectAll('option')
  .data(['flare', 'simple']).enter()
  .append('option')
  .text(function (d) { return d })

function onchange () {
  const selectValue = d3.select('select').property('value')
  console.log('selected ', selectValue)
  load('data/' + selectValue + '.json')
  // d3.select('body')
  //   .append('p')
  //   .text(selectValue + ' is the last selected option.')
};

onchange() // load('flare.json')
d3.select(window.frameElement).style('height', height + 'px')
