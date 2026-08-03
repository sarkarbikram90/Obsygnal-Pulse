from setuptools import setup, find_packages

setup(
    name="obsygnal-pulse",
    version="0.1.0",
    description="Python SDK for Obsygnal-Pulse AI Infrastructure Observability & Optimization Engine",
    author="Obsygnal Labs",
    author_email="engineering@obsygnal.io",
    packages=find_packages(),
    install_requires=[
        "requests>=2.25.0",
    ],
    python_requires=">=3.8",
    classifiers=[
        "Programming Language :: Python :: 3",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
        "Topic :: System :: Monitoring",
    ],
)
