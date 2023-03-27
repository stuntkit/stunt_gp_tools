#!/usr/bin/env python3
# coding=utf-8

import os
import sys
import argparse

import struct
from networkx import nx  # type: ignore

parser = argparse.ArgumentParser(formatter_class=argparse.ArgumentDefaultsHelpFormatter)
parser.add_argument(
    "filename", help="name of the save file", nargs="?", type=str, default="setup.bin"
)
parser.add_argument(
    "-n",
    "--count",
    help="ncount of best tracks to diplay",
    nargs="?",
    type=int,
    default=3,
)


class PointTime:
    """Stores Name and time in the way the game stores it"""

    def __init__(self, time: int, name: str = "..."):
        self.name = name
        self.time = time

    def __str__(self):
        minutes = int(self.time / 60 / 60)
        seconds = int(self.time / 60 - (minutes * 60))
        milliseconds = int((self.time / 60 - (minutes * 60) - seconds) * 100)
        return f"{minutes:02d}:{seconds:02d}.{milliseconds:02d}"


class PointScore:
    """Stores Name and score"""

    def __init__(self, score: int = 0, name: str = "..."):
        self.name = name
        self.score = score


class ArcadePoint:
    """stores arcade data"""

    def __init__(self):
        # default time for "..." player, 2 mins
        self.lap = PointTime(2 * 60 * 60, "...")
        # 10 minutes, total race time
        self.total = PointTime(10 * 60 * 60, "...")
        self.unknown = PointScore(0, "...")
        self.stunts = PointScore(0, "...")


def get_name(name: bytes) -> str:
    """converts name to ascii string"""
    terminator = name.index(b"\x00")
    return name[:terminator].decode("ascii")


def add_edge_if(graph, start: str, end: str, edge_weight: int):
    """adds edge in the graph if the time is not 0"""
    if edge_weight is None:
        graph.add_edge(start, end, weight=edge_weight)


def convert_time(time: int) -> str:
    """converts time from the game's internal format"""
    minutes = int(time / 60 / 60)
    seconds = int(time / 60 - (minutes * 60))
    milliseconds = int((time / 60 - (minutes * 60) - seconds) * 100)
    return f"{minutes:02d}:{seconds:02d}.{milliseconds:02d}"


def main():
    """Finds fastest Arcade route for speedrunning"""
    args = parser.parse_args()
    filename: str = args.filename

    points = [None] * 18
    if not os.path.exists(filename) or not os.path.isfile(filename):
        print(f"File {filename} not found", file=sys.stderr)
    else:
        # deepcode ignore PT: cli tool
        with open(filename, "rb") as save_file:
            # skip first one, because it's unused
            save_file.seek(0x114 + 0x20)
            for i in range(18):
                point: ArcadePoint = ArcadePoint()

                point.lap.time = struct.unpack("<L", save_file.read(4))[0]
                point.lap.name = get_name(save_file.read(4))
                point.total.time = struct.unpack("<L", save_file.read(4))[0]
                point.total.name = get_name(save_file.read(4))
                point.unknown.score = struct.unpack("<L", save_file.read(4))[0]
                point.unknown.name = get_name(save_file.read(4))
                point.stunts.score = struct.unpack("<L", save_file.read(4))[0]
                point.stunts.name = get_name(save_file.read(4))

                if point.total.name != b"...":
                    # add to list if the point is not empty
                    points[i] = point

        g = generate_graph(points)

        # there are 24 paths max from start to end
        shortests = list(nx.shortest_simple_paths(g, "start", "end", "weight"))
        if len(shortests) < 0:
            print("cannot find a path from start to end", file=sys.stderr)
        else:
            if len(shortests) < args.count:
                print(
                    f"found only {len(shortests)} shortest paths,"
                    " but {args.count} were requested",
                    file=sys.stderr,
                )
                # for insurance everything will work fine
                args.count = len(shortests)
            # n fastest routes
            shortests = shortests[: args.count]

            for i, path in enumerate(shortests, start=1):
                times = []
                total = 0
                for current_point in range(1, len(path) - 1):
                    time = g.get_edge_data(
                        path[current_point], path[current_point + 1]
                    )["weight"]
                    total += time
                    time_string = convert_time(time)
                    times.append(time_string)
                print(f"Path #{i}:")
                # cut start and end points
                print(" -> ".join(path[1:-1]))
                print(f'{" + ".join(times)} = {convert_time(total)}')
                print()

            print()
            print(r"08    05    17    21    14    10")
            print(r"  \  /  \  /  \  /  \  /  \  /  ")
            print(r"   19    18    15    16    01   ")
            print(r"     \  /  \  /  \  /  \  /     ")
            print(r"      07    22    09    04      ")
            print(r"        \  /  \  /  \  /        ")
            print(r"         20    06    02         ")
            print()
            print(
                "See https://github.com/Halamix2/stunt_gp_formats/wiki/Tracks"
                " to see what track names corresponds to each ID"
            )


def generate_graph(points):
    """generates the connection graph and fills it with the drive times"""
    # check if weight can be assigned to nodes rather than edges
    g = nx.DiGraph()
    g.add_nodes_from(
        [
            "start",
            "01",
            "02",
            "04",
            "05",
            "06",
            "07",
            "08",
            "09",
            "10",
            "14",
            "15",
            "16",
            "17",
            "18",
            "19",
            "20",
            "21",
            "22",
            "end",
        ]
    )
    # 6 -> end points connections
    add_edge_if(g, "08", "end", points[0].total.time)
    add_edge_if(g, "05", "end", points[1].total.time)
    add_edge_if(g, "17", "end", points[2].total.time)
    add_edge_if(g, "21", "end", points[3].total.time)
    add_edge_if(g, "14", "end", points[4].total.time)
    add_edge_if(g, "10", "end", points[5].total.time)

    # 5 -> 6 points connections
    add_edge_if(g, "19", "08", points[6].total.time)
    add_edge_if(g, "19", "05", points[6].total.time)
    add_edge_if(g, "18", "05", points[7].total.time)
    add_edge_if(g, "18", "17", points[7].total.time)
    add_edge_if(g, "15", "17", points[8].total.time)
    add_edge_if(g, "15", "21", points[8].total.time)
    add_edge_if(g, "16", "21", points[9].total.time)
    add_edge_if(g, "16", "14", points[9].total.time)
    add_edge_if(g, "01", "14", points[10].total.time)
    add_edge_if(g, "01", "10", points[10].total.time)

    # 4 -> 5 points connections
    add_edge_if(g, "07", "19", points[11].total.time)
    add_edge_if(g, "07", "18", points[11].total.time)
    add_edge_if(g, "22", "18", points[12].total.time)
    add_edge_if(g, "22", "15", points[12].total.time)
    add_edge_if(g, "09", "15", points[13].total.time)
    add_edge_if(g, "09", "16", points[13].total.time)
    add_edge_if(g, "04", "16", points[14].total.time)
    add_edge_if(g, "04", "01", points[14].total.time)

    # 3 -> 4 points connections
    add_edge_if(g, "20", "07", points[15].total.time)
    add_edge_if(g, "20", "22", points[15].total.time)
    add_edge_if(g, "06", "22", points[16].total.time)
    add_edge_if(g, "06", "09", points[16].total.time)
    add_edge_if(g, "02", "09", points[17].total.time)
    add_edge_if(g, "02", "04", points[17].total.time)

    # start -> 3 points connections
    add_edge_if(g, "start", "20", 0)
    add_edge_if(g, "start", "06", 0)
    add_edge_if(g, "start", "02", 0)

    return g


# in the future: print only unlocked paths (\/)
# in the future: another script for showing what was finished?
# in the future: prepare for new modes, all tracks and all connections speedruns
if __name__ == "__main__":
    main()
